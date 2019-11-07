package authz

import (
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/emirpasic/gods/maps/hashmap"
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	menshenext "github.com/it2911/menshen/pkg/ext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var UserMap = hashmap.New()
var RoleBindingExtMap = hashmap.New()
var RoleExtMap = hashmap.New()
var GroupMap = hashmap.New()

type RoleBindingInfo struct {
	RoleExtNames []string
	Type         string
	Message      string
}

func Cache() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	// BuildConfigFromFlags is a helper function that builds configs from a master url or
	// a kubeconfig filepath.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	logerror(err)

	clientset, err := menshenext.NewForConfig(config)
	logerror(err)

	roleExtList, err := clientset.RoleExts().List(metav1.ListOptions{})
	logerror(err)

	for _, roleExt := range roleExtList.Items {
		roleName := roleExt.GetName()
		for _, role := range roleExt.Spec.Roles {
			verbMap := hashmap.New()
			for _, verb := range role.Verbs {
				resourceMap := hashmap.New()
				for _, resource := range role.Resources {
					resourceNameMap := hashmap.New()
					for _, resourceName := range role.ResourceNames {
						apiGroupMap := hashmap.New()
						for _, apiGroup := range role.ApiGroups {
							namespaceMap := hashmap.New()
							for _, namespace := range role.Namespaces {
								namespaceMap.Put(namespace, "")
							}
							apiGroupMap.Put(apiGroup, namespaceMap)
						}
						resourceNameMap.Put(resourceName, apiGroupMap)
					}
					resourceMap.Put(resource, resourceNameMap)
				}
				for _, nonresource := range role.NonResources {
					resourceMap.Put(nonresource, "")
				}
				verbMap.Put(verb, resourceMap)
			}
			RoleExtMap.Put(roleName, verbMap)
		}
	}

	groupextList := &menshenv1beta1.GroupExtList{}
	groupextList, err = clientset.GroupExts().List(metav1.ListOptions{})
	logerror(err)
	for _, groupExt := range groupextList.Items {
		GroupMap.Put(groupExt.Name, groupExt.Spec.Users)
	}

	rolebindingextList, err := clientset.RoleBindingExts().List(metav1.ListOptions{})
	logerror(err)
	for _, rolebindingext := range rolebindingextList.Items {

		for _, subject := range rolebindingext.Spec.Subjects {
			if strings.EqualFold(subject.Kind, "User") ||
				strings.EqualFold(subject.Kind, "ServiceAccount") {

				if rolebindingExtNameList, found := UserMap.Get(subject.Name); !found {
					rolebindingExtNameList.(*arraylist.List).Add(rolebindingext.Name)
					UserMap.Put(subject.Name, rolebindingExtNameList)
				} else {
					rolebindingExtNameList.(*arraylist.List).Add(rolebindingext.Name)
				}
			} else if strings.EqualFold(subject.Kind, "Group") {
				users, found := GroupMap.Get(subject.Name)
				if found {
					for _, user := range users.([]string) {
						if rolebindingExtNameList, found := UserMap.Get(user); !found {
							rolebindingExtNameList.(*arraylist.List).Add(rolebindingext.Name)
							UserMap.Put(user, rolebindingExtNameList)
						} else {
							rolebindingExtNameList.(*arraylist.List).Add(rolebindingext.Name)
						}
					}
				}
			}
		}

		RoleBindingExtMap.Put(rolebindingext.Name,
			RoleBindingInfo{RoleExtNames: rolebindingext.Spec.RoleNames, Type: rolebindingext.Spec.Type, Message: rolebindingext.Spec.Message})
	}
}

func logerror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
