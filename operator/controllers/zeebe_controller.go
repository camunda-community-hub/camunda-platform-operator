/*
Copyright 2021.

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

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	camundacloudv1 "io.camnda/operator/api/v1"
)

// ZeebeReconciler reconciles a Zeebe object
type ZeebeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=camunda-cloud.io.camunda,resources=zeebes/finalizers,verbs=update

// CRUD apps: deployments and statefulsets
// +kubebuilder:rbac:groups=apps,resources=statefulsets;deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status;deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Zeebe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ZeebeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var zeebe camundacloudv1.Zeebe
	if err := r.Get(ctx, req.NamespacedName, &zeebe); err != nil {
		log.Error(err, "unable to fetch Statefulset")

		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	brokerStatefulSet := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        "Zeebe",
			Namespace:   req.Namespace,
		},
	}

	if err := ctrl.SetControllerReference(&zeebe, brokerStatefulSet, r.Scheme); err != nil {
		if err != nil {
			log.Error(err, "unable to construct statefulset from zeebe CRD")
			// don't bother requeuing until we get a change to the spec
			return ctrl.Result{}, nil
		}
	}

	// We return an empty result and no error,
	// which indicates to controller-runtime that we’ve successfully reconciled
	// this object and don’t need to try again until there’s some changes.
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZeebeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&camundacloudv1.Zeebe{}).
		Complete(r)
}
