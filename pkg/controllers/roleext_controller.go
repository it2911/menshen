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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
)

var RoleExts []authv1beta1.Role

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

	// your logic here

	return ctrl.Result{}, nil
}

func (r *RoleExtReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1beta1.RoleExt{}).
		Complete(r)
}