package authz

import (
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/hashset"
	authv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

var RoleBindingMap hashmap.Map

func cache() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	// BuildConfigFromFlags is a helper function that builds configs from a master url or
	// a kubeconfig filepath.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	logerror(err)

	clientset, err := authv1beta1.NewForConfig(config)
	logerror(err)

	groupextList := &authv1beta1.GroupExtList{}
	groupextList, err = clientset.GroupExts().List(metav1.ListOptions{})
	logerror(err)

	groupMap := make(map[string][]string)
	for _, groupext := range groupextList.Items {
		groupMap[groupext.Name] = groupext.Spec.Users
	}

	rolebindingextList := &authv1beta1.RoleBindingExtList{}
	rolebindingextList, err = clientset.RoleBindingExts().List(metav1.ListOptions{})
	logerror(err)

	roleExtMap := hashmap.New()
	roleMap := hashmap.New()
	for _, rolebindingext := range rolebindingextList.Items {
		for _, binding := range rolebindingext.Spec.Bindings {
			userSet := hashset.New()
			for _, subject := range binding.Subjects {
				if subject.Kind == "User" || subject.Kind == "ServiceAccount" {
					userSet.Add(subject.Name)
				} else if subject.Kind == "Group" {
					users := groupMap[subject.Name]
					if users != nil {
						for user := range users {
							userSet.Add(user)
						}
					}
				}
			}
			for _, roleName := range binding.RoleNames {
				var namespaceMap *hashmap.Map
				namespaceMap = hashmap.New()
				roleExtMap.Put(roleName, map[string]string{"type": binding.Type, "message": binding.Message})
				RoleBindingMap.Put(userSet, namespaceMap)
			}
		}
	}

	roleextList := &authv1beta1.RoleExtList{}
	roleextList, err = clientset.RoleExts().List(metav1.ListOptions{})
	logerror(err)

	for _, roleext := range roleextList.Items {
		roleName := roleext.GetName()
		nsmap, found := roleMap.Get(roleName)
		var namespaceMap *hashmap.Map

		if found {
			namespaceMap = nsmap.(*hashmap.Map)
		} else {
			namespaceMap = hashmap.New()
			roleMap.Put(roleName, namespaceMap)
		}

		for _, role := range roleext.Spec.Roles {
			for _, namespace := range role.Namespaces {
				if namespace == "*" || namespace == "" {
					namespace = "_"
				}
				apigroupMap := hashmap.New()
				for _, apigroup := range role.ApiGroups {
					if apigroup == "*" || apigroup == "" {
						apigroup = "_"
					}
					verbMap := hashmap.New()
					for _, verb := range role.Verbs {
						if verb == "*" || verb == "" {
							verb = "_"
						}
						resourceMap := hashmap.New()
						for _, resource := range role.Resources {
							if resource == "*" || resource == "" {
								resource = "_"
							}
							resourceNameMap := hashmap.New()
							for _, resourceName := range role.ResourceNames {
								if resourceName == "*" || resourceName == "" {
									resourceName = "_"
								}
								resourceNameMap.Put(resourceName, "")
							}
							resourceMap.Put(resource, resourceNameMap)
						}
						for _, nonresource := range role.NonResources {
							if nonresource == "*" || nonresource == "" {
								nonresource = "_"
							}
							resourceMap.Put(nonresource, "")
						}
						verbMap.Put(verb, resourceMap)
					}
					apigroupMap.Put(apigroup, verbMap)
				}
				namespaceMap.Put(namespace, apigroupMap)
			}
		}
	}
}

func logerror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
