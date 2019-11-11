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
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/it2911/menshen/pkg/ext"
	"github.com/it2911/menshen/pkg/utils"
	"k8s.io/klog"
	"strings"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
)

var UserBindingMap = hashmap.New()
var GroupBindingMap = hashmap.New()
var RoleBindingExtAllowMap = hashmap.New()
var RoleBindingExtDenyMap = hashmap.New()

type RoleBindingExtInfo struct {
	RoleExtNames []string
	Message      string
}

// RoleBindingExtReconciler reconciles a RoleBindingExt object
type RoleBindingExtReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=auth.menshen.io,resources=rolebindingexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auth.menshen.io,resources=rolebindingexts/status,verbs=get;update;patch

func (r *RoleBindingExtReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("rolebindingext", req.NamespacedName)

	roleBindingExtName := req.Name
	roleBindingExt, err := ext.MenShenClientSet.RoleBindingExts().Get(req.Name, req.Namespace)

	if err != nil {
		if _, found := RoleBindingExtAllowMap.Get(roleBindingExtName); found {
			RoleBindingExtAllowMap.Remove(roleBindingExtName)
			klog.Infof("RoleBindingExtAllowMap remove %s .", roleBindingExtName)
		}

		if _, found := RoleBindingExtDenyMap.Get(roleBindingExtName); found {
			RoleBindingExtDenyMap.Remove(roleBindingExtName)
			klog.Infof("RoleBindingExtDenyMap remove %s .", roleBindingExtName)
		}

		groupBindingKeys := GroupBindingMap.Keys()
		for _, key := range groupBindingKeys {
			roleBindingExtNameListI, found := GroupBindingMap.Get(key)
			if found {
				roleBindingExtNameList := roleBindingExtNameListI.(*hashset.Set)
				roleBindingExtNameList.Remove(roleBindingExtName)

				logString, err := json.Marshal(roleBindingExtNameList)
				utils.HandleErr(err)
				klog.Infof("GroupBindingMap ％s remove %s, left %s .", key, roleBindingExtName, logString)

				if roleBindingExtNameList.Size() == 0 {
					GroupBindingMap.Remove(key)
					klog.Infof("%s is empty. Removed from GroupBindingMap.", key)
				}
			}
		}

		userBindingKeys := UserBindingMap.Keys()
		for _, key := range userBindingKeys {
			roleBindingExtNameListI, found := UserBindingMap.Get(key)
			if found {
				roleBindingExtNameList := roleBindingExtNameListI.(*hashset.Set)
				roleBindingExtNameList.Remove(roleBindingExtName)

				logString, err := json.Marshal(roleBindingExtNameList)
				utils.HandleErr(err)
				klog.Infof("GroupBindingMap ％s remove %s, left %s .", key, roleBindingExtName, logString)

				if roleBindingExtNameList.Size() == 0 {
					UserBindingMap.Remove(key)
					klog.Infof("%s is empty. Removed from GroupBindingMap.", key)
				}
			}
		}
	} else {
		for _, subject := range roleBindingExt.Spec.Subjects {
			if strings.EqualFold(subject.Kind, "User") ||
				strings.EqualFold(subject.Kind, "ServiceAccount") {

				roleBindingExtNameList := hashset.New()
				roleBindingExtNameListI, found := UserBindingMap.Get(subject.Name)
				if found {
					roleBindingExtNameList = roleBindingExtNameListI.(*hashset.Set)
					roleBindingExtNameList.Add(roleBindingExtName)
					klog.Infof("UserBinding Map %s add %s into its binding list.", subject.Name, roleBindingExtName)
				} else {
					roleBindingExtNameList.Add(roleBindingExtName)
					klog.Infof("UserBinding Map %s add %s into its binding list.", subject.Name, roleBindingExtName)
				}
				UserBindingMap.Put(subject.Name, roleBindingExtNameList)
			} else if strings.EqualFold(subject.Kind, "Group") {

				roleBindingExtNameList := hashset.New()
				roleBindingExtNameListI, found := GroupBindingMap.Get(subject.Name)
				if found {
					roleBindingExtNameList = roleBindingExtNameListI.(*hashset.Set)
					roleBindingExtNameList.Add(roleBindingExtName)
					klog.Infof("GroupBinding Map %s add %s into its binding list.", subject.Name, roleBindingExtName)
				} else {
					roleBindingExtNameList.Add(roleBindingExtName)
					klog.Infof("GroupBinding Map %s add %s into its binding list.", subject.Name, roleBindingExtName)
				}
				GroupBindingMap.Put(subject.Name, roleBindingExtNameList)
			}
		}

		if strings.EqualFold(roleBindingExt.Spec.Type, "allow") {
			RoleBindingExtAllowMap.Put(roleBindingExtName, RoleBindingExtInfo{RoleExtNames: roleBindingExt.Spec.RoleNames})
			logString, err := json.Marshal(roleBindingExt.Spec.RoleNames)
			utils.HandleErr(err)
			klog.Infof("RoleBindingExtAllowMap %s add role list %s.", roleBindingExtName, logString)
		} else if strings.EqualFold(roleBindingExt.Spec.Type, "deny") {
			RoleBindingExtDenyMap.Put(roleBindingExtName, RoleBindingExtInfo{RoleExtNames: roleBindingExt.Spec.RoleNames, Message: roleBindingExt.Spec.Message})
			logString, err := json.Marshal(roleBindingExt.Spec.RoleNames)
			utils.HandleErr(err)
			klog.Infof("RoleBindingExtDenyMap %s add role list %s .", roleBindingExtName, logString)
		} else {
			// TODO
			klog.Error("Without type:" + roleBindingExtName)
		}
	}

	return ctrl.Result{}, nil
}

func (r *RoleBindingExtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&menshenv1beta1.RoleBindingExt{}).
		Complete(r)
}
