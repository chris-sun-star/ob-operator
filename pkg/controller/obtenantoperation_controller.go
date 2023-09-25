/*
Copyright 2023.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oceanbasev1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
)

// OBTenantOperationReconciler reconciles a OBTenantOperation object
type OBTenantOperationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantoperations/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenant,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenant/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *OBTenantOperationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = req
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantOperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&oceanbasev1alpha1.OBTenantOperation{}).
		Complete(r)
}
