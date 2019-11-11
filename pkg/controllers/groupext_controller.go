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

var GroupUserMap = hashmap.New()

// GroupExtReconciler reconciles a GroupExt object
type GroupExtReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=auth.menshen.io,resources=groupexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auth.menshen.io,resources=groupexts/status,verbs=get;update;patch

func (r *GroupExtReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("groupext", req.NamespacedName)

	groupName := req.Name
	groupExt, err := ext.MenShenClientSet.GroupExts().Get(req.Name)
	utils.HandleErr(err)

	if groupExt.UID == "" {
		GroupUserMap.Remove(groupName)
		klog.Infof("Group Map remove %s group", groupName)
	} else {
		GroupUserMap.Put(groupName, groupExt.Spec.Users)
		logString, err := json.Marshal(groupExt.Spec.Users)
		utils.HandleErr(err)
		klog.Infof("Group Map add %s group, user list %s", groupName, string(logString))
	}

	return ctrl.Result{}, err
}

func (r *GroupExtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&menshenv1beta1.GroupExt{}).
		Complete(r)
}
