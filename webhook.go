package main

import (
	"encoding/json"
	"fmt"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/gertd/go-pluralize"
	"github.com/it2911/menshen/pkg/controllers"
	"github.com/it2911/menshen/pkg/utils"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"

	"github.com/golang/glog"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	requiredLabels = []string{
		nameLabel,
		instanceLabel,
		versionLabel,
		componentLabel,
		partOfLabel,
		managedByLabel,
	}
	addLabels = map[string]string{
		nameLabel:      NA,
		instanceLabel:  NA,
		versionLabel:   NA,
		componentLabel: NA,
		partOfLabel:    NA,
		managedByLabel: NA,
	}
)

const (
	admissionWebhookAnnotationValidateKey = "admission-webhook-example.banzaicloud.com/validate"
	admissionWebhookAnnotationMutateKey   = "admission-webhook-example.banzaicloud.com/mutate"
	admissionWebhookAnnotationStatusKey   = "admission-webhook-example.banzaicloud.com/status"

	nameLabel      = "app.kubernetes.io/name"
	instanceLabel  = "app.kubernetes.io/instance"
	versionLabel   = "app.kubernetes.io/version"
	componentLabel = "app.kubernetes.io/component"
	partOfLabel    = "app.kubernetes.io/part-of"
	managedByLabel = "app.kubernetes.io/managed-by"

	NA = "not_available"
)

type WebhookServer struct {
	server *http.Server
}

// Webhook Server parameters
type WhSvrParameters struct {
	port           int    // webhook server port
	certFile       string // path to the x509 certificate for https
	keyFile        string // path to the x509 private key matching `CertFile`
	sidecarCfgFile string // path to sidecar injector configuration file
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = corev1.AddToScheme(runtimeScheme)
}

func admissionRequired(ignoredList []string, admissionAnnotationKey string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var required bool
	switch strings.ToLower(annotations[admissionAnnotationKey]) {
	default:
		required = true
	case "n", "no", "false", "off":
		required = false
	}
	return required
}

func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, admissionWebhookAnnotationMutateKey, metadata)
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	status := annotations[admissionWebhookAnnotationStatusKey]

	if strings.ToLower(status) == "mutated" {
		required = false
	}

	glog.Infof("Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, admissionWebhookAnnotationValidateKey, metadata)
	glog.Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func updateAnnotation(target map[string]string, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}

func updateLabels(target map[string]string, added map[string]string) (patch []patchOperation) {
	values := make(map[string]string)
	for key, value := range added {
		if target == nil || target[key] == "" {
			values[key] = value
		}
	}
	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: values,
	})
	return patch
}

func createPatch(availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	patch = append(patch, updateLabels(availableLabels, labels)...)

	return json.Marshal(patch)
}

// validate deployments and services
func (whsvr *WebhookServer) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var (
		availableLabels                 map[string]string
		objectMeta                      *metav1.ObjectMeta
		resourceNamespace, resourceName string
	)

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
		availableLabels = deployment.Labels
	case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
	}

	if !validationRequired(ignoredNamespaces, objectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	allowed := true
	var result *metav1.Status
	glog.Info("available labels:", availableLabels)
	glog.Info("required labels", requiredLabels)
	for _, rl := range requiredLabels {
		if _, ok := availableLabels[rl]; !ok {
			allowed = false
			result = &metav1.Status{
				Reason: "required labels are not set",
			}
			break
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: allowed,
		Result:  result,
	}
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var (
		availableLabels, availableAnnotations map[string]string
		objectMeta                            *metav1.ObjectMeta
		resourceNamespace, resourceName       string
	)

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)

	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
		availableLabels = deployment.Labels
	case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
	}

	if !mutationRequired(ignoredNamespaces, objectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", resourceNamespace, resourceName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{admissionWebhookAnnotationStatusKey: "mutated"}
	patchBytes, err := createPatch(availableAnnotations, annotations, availableLabels, addLabels)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook server
func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/mutate" {
			admissionResponse = whsvr.mutate(&ar)
		} else if r.URL.Path == "/validate" {
			admissionResponse = whsvr.validate(&ar)
		} else if r.URL.Path == "/authorize" {

		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	scheme := runtime.NewScheme()
	codecFactory := serializer.NewCodecFactory(scheme)
	deserializer := codecFactory.UniversalDeserializer()

	sarObject, _, err := deserializer.Decode(body, nil, &authorizationv1.SubjectAccessReview{})
	utils.HandleErr(err)

	sar := sarObject.(*authorizationv1.SubjectAccessReview)
	var sarRespState authorizationv1.SubjectAccessReviewStatus
	username := sar.Spec.User

	// Allow Role Check
	if rolebindingExtNameListI, found := controllers.UserBindingMap.Get(sar.Spec.User); found {
		sarRespState, found = CheckAllowRole(rolebindingExtNameListI.(*hashset.Set), sar)
		if found {
			sar.Status = sarRespState
			SubjectAccessReviewResponse(w, sar)
			return
		}
	}

	for _, groupName := range controllers.GroupBindingMap.Keys() {
		users, found := controllers.GroupUserMap.Get(groupName)
		if found {
			for _, user := range users.([]string) {
				if user == username {
					if roleBindingExtNameListI, found := controllers.GroupBindingMap.Get(groupName); found {
						sarRespState, found = CheckAllowRole(roleBindingExtNameListI.(*hashset.Set), sar)
						if found {
							sar.Status = sarRespState
							SubjectAccessReviewResponse(w, sar)
							return
						}
					}
				}
			}
		}
	}

	// Deny Role Check
	if rolebindingExtNameListI, found := controllers.UserBindingMap.Get(sar.Spec.User); found {
		sarRespState, found = CheckDenyRole(rolebindingExtNameListI.(*hashset.Set), sar)
		if found {
			sar.Status = sarRespState
			SubjectAccessReviewResponse(w, sar)
			return
		}
	}

	for _, groupName := range controllers.GroupBindingMap.Keys() {
		users, found := controllers.GroupUserMap.Get(groupName)
		if found {
			for _, user := range users.([]string) {
				if user == username {
					if rolebindingExtNameListI, found := controllers.GroupBindingMap.Get(groupName); found {
						if sarRespState, found = CheckDenyRole(rolebindingExtNameListI.(*hashset.Set), sar); found {
							sar.Status = sarRespState
							SubjectAccessReviewResponse(w, sar)
							return
						}
					}
				}
			}
		}
	}

	sar.Status = authorizationv1.SubjectAccessReviewStatus{Allowed: true}
	SubjectAccessReviewResponse(w, sar)
}

func resourceRoleFound(sar *authorizationv1.SubjectAccessReview, roleExt controllers.RoleExtInfo) bool {

	found := true
	pluralize := pluralize.NewClient()

	if resourceAttributes := sar.Spec.ResourceAttributes; resourceAttributes != nil {

		if resourceAttributes.Namespace == "" {
			resourceAttributes.Namespace = "_"
		}

		var verbMapI interface{}
		if verbMapI, found = roleExt.ResourceMap.Get(resourceAttributes.Namespace); !found {
			if verbMapI, found = roleExt.ResourceMap.Get("*"); !found {
				return found
			}
		}

		var apiGroupMapI interface{}
		verbMap := verbMapI.(*hashmap.Map)
		if apiGroupMapI, found = verbMap.Get(resourceAttributes.Verb); !found {
			if apiGroupMapI, found = verbMap.Get("*"); !found {
				return found
			}
		}

		apiGroup := resourceAttributes.Version
		apiGroupMap := apiGroupMapI.(*hashmap.Map)
		var resourceMapI interface{}
		if resourceAttributes.Group != "" {
			apiGroup = resourceAttributes.Group + "/" + apiGroup
		}
		if resourceMapI, found = apiGroupMap.Get(apiGroup); !found {
			if resourceMapI, found = apiGroupMap.Get("*"); !found {
				return found
			}
		}

		var resourceNameMapI interface{}
		resourceMap := resourceMapI.(*hashmap.Map)
		var singular, plural string
		if pluralize.IsPlural(resourceAttributes.Resource) {
			plural = resourceAttributes.Resource
			singular = pluralize.Singular(resourceAttributes.Resource)
		} else if pluralize.IsSingular(resourceAttributes.Resource) {
			plural = pluralize.Plural(resourceAttributes.Resource)
			singular = resourceAttributes.Resource
		} else {
			plural = resourceAttributes.Resource
			singular = resourceAttributes.Resource
		}
		if resourceNameMapI, found = resourceMap.Get(singular); !found {
			if resourceNameMapI, found = resourceMap.Get(plural); !found {
				if resourceNameMapI, found = resourceMap.Get("*"); !found {
					return found
				}
			}
		}

		resourceNameMap := resourceNameMapI.(*hashmap.Map)
		if _, found = resourceNameMap.Get(resourceAttributes.Name); !found {
			if _, found = resourceNameMap.Get("*"); !found {
				return found
			}
		}
	} else if nonResourceAttributes := sar.Spec.NonResourceAttributes; nonResourceAttributes != nil {

		var resourceMapI interface{}
		if resourceMapI, found = roleExt.NonResourceMap.Get(nonResourceAttributes.Verb); !found {
			if resourceMapI, found = roleExt.NonResourceMap.Get("*"); !found {
				return found
			}
		}

		resourceMap := resourceMapI.(*hashmap.Map)
		if _, found := resourceMap.Get("*"); !found {
			if _, found = resourceMap.Get(nonResourceAttributes.Path); !found {
				return found
			}
		}
	}

	return found
}

func CheckAllowRole(roleBindingExtNameList *hashset.Set, sar *authorizationv1.SubjectAccessReview) (authorizationv1.SubjectAccessReviewStatus, bool) {
	for _, roleBindingName := range roleBindingExtNameList.Values() {
		// Allow List Check
		roleBindingExtI, found := controllers.RoleBindingExtAllowMap.Get(roleBindingName)
		if found {
			roleBindingExt := roleBindingExtI.(controllers.RoleBindingExtInfo)
			for _, roleExtName := range roleBindingExt.RoleExtNames {
				roleExt, found := controllers.RoleExtMap.Get(roleExtName)
				if found {
					found = resourceRoleFound(sar, roleExt.(controllers.RoleExtInfo))
					if found {
						return authorizationv1.SubjectAccessReviewStatus{Allowed: true}, true
					}
				}
			}
		}
	}

	// Haven't Allow/Deny Rule
	return authorizationv1.SubjectAccessReviewStatus{}, false
}

func CheckDenyRole(roleBindingExtNameList *hashset.Set, sar *authorizationv1.SubjectAccessReview) (authorizationv1.SubjectAccessReviewStatus, bool) {
	for _, roleBindingName := range roleBindingExtNameList.Values() {
		// Deny List Check
		roleBindingExtI, found := controllers.RoleBindingExtDenyMap.Get(roleBindingName)
		if found {
			roleBindingExt := roleBindingExtI.(controllers.RoleBindingExtInfo)
			for _, roleExtName := range roleBindingExt.RoleExtNames {
				roleExt, found := controllers.RoleExtMap.Get(roleExtName)
				if found {
					found = resourceRoleFound(sar, roleExt.(controllers.RoleExtInfo))
					if found {
						return authorizationv1.SubjectAccessReviewStatus{Denied: true, Reason: roleBindingExt.Message}, true
					}
				}
			}
		}
	}

	// Haven't Allow/Deny Rule
	return authorizationv1.SubjectAccessReviewStatus{}, false
}

func SubjectAccessReviewResponse(w http.ResponseWriter, sar *authorizationv1.SubjectAccessReview) {

	if resp, err := json.Marshal(sar); err != nil {
		utils.HandleErr(err)
	} else {
		klog.Info(string(resp))

		if _, err = w.Write(resp); err != nil {
			klog.Errorf("Can't write response: %v", err)
			http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		}
	}
}
