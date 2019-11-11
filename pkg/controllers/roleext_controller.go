/*
Copyright 2019 chengchen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/go-logr/logr"
	"github.com/it2911/menshen/pkg/ext"
	"github.com/it2911/menshen/pkg/utils"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
)

var RoleExtMap = hashmap.New()

var RoleExts []menshenv1beta1.Role

type RoleExtInfo struct {
	ResourceMap    *hashmap.Map
	NonResourceMap *hashmap.Map
}

// RoleExtReconciler reconciles a RoleExt object
type RoleExtReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=auth.menshen.io,resources=roleexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auth.menshen.io,resources=roleexts/status,verbs=get;update;patch
func (r *RoleExtReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("roleext", req.NamespacedName)

	roleName := req.Name
	roleExt, err := ext.MenShenClientSet.RoleExts().Get(req.Name)
	if err != nil {
		klog.Error(err)
	}

	if roleExt.UID == "" {
		RoleExtMap.Remove(roleName)
		klog.Infof("RoleExt Map remove %s .", roleName)
	} else {
		roleMap := GetRoleExtInfo(roleExt.Spec.Roles)
		RoleExtMap.Put(roleName, roleMap)

		logString, err := json.Marshal(roleExt.Spec.Roles)
		utils.HandleErr(err)
		klog.Infof("Role Map %s origin list %s", roleName, string(logString))
		logString, err = json.Marshal(roleMap)
		utils.HandleErr(err)
		klog.Infof("Role Map %s convert list %s", roleName, string(logString))
	}

	return ctrl.Result{}, nil
}

func (r *RoleExtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&menshenv1beta1.RoleExt{}).
		Complete(r)
}

func GetRoleExtInfo(roles []menshenv1beta1.Role) RoleExtInfo {
	roleExtInfoResourceMap := hashmap.New()
	roleExtInfoNonResourceMap := hashmap.New()

	for _, role := range roles {
		if role.Resources != nil {
			if role.Namespaces == nil {
				role.Namespaces = append(role.Namespaces, "_")
			}
			for _, namespace := range role.Namespaces {
				if namespace == "" {
					namespace = "*"
				}
				verbMap := hashmap.New()
				verbMapI, found := roleExtInfoResourceMap.Get(namespace)
				if found {
					verbMap = verbMapI.(*hashmap.Map)
				}
				for _, verb := range role.Verbs {
					if verb == "" {
						verb = "*"
					}
					apiGroupMap := hashmap.New()
					apiGroupMapI, found := verbMap.Get(verb)
					if found {
						apiGroupMap = apiGroupMapI.(*hashmap.Map)
					}
					for _, apiGroup := range role.ApiGroups {
						if apiGroup == "" {
							apiGroup = "v1"
						}
						resourceMap := hashmap.New()
						resourceMapI, found := apiGroupMap.Get(apiGroup)
						if found {
							resourceMap = resourceMapI.(*hashmap.Map)
						}

						for _, resource := range role.Resources {
							if resource == "" {
								resource = "*"
							}
							resourceNameMap := hashmap.New()
							resourceNameMapI, found := resourceNameMap.Get(resource)
							if found {
								resourceNameMap = resourceNameMapI.(*hashmap.Map)
							}
							if role.ResourceNames == nil {
								role.ResourceNames = append(role.ResourceNames, "*")
							}
							for _, resourceName := range role.ResourceNames {
								if resourceName == "" {
									resourceName = "*"
								}
								resourceNameMap.Put(resourceName, "")
							}
							resourceMap.Put(resource, resourceNameMap)
						}
						apiGroupMap.Put(apiGroup, resourceMap)
					}
					verbMap.Put(verb, apiGroupMap)
				}
				roleExtInfoResourceMap.Put(namespace, verbMap)
			}
		} else if role.NonResources != nil {
			for _, verb := range role.Verbs {
				nonResourceMap := hashmap.New()
				nonResourceMapI, found := roleExtInfoNonResourceMap.Get(verb)
				if found {
					nonResourceMap = nonResourceMapI.(*hashmap.Map)
				}
				for _, nonresource := range role.NonResources {
					nonResourceMap.Put(nonresource, "")
				}
				if verb == "" {
					verb = "*"
				}
				nonResourceMap.Put(verb, nonResourceMap)
			}
		}
	}

	return RoleExtInfo{ResourceMap: roleExtInfoResourceMap, NonResourceMap: roleExtInfoNonResourceMap}
}
