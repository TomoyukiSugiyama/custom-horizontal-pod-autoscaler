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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

// CustomHorizontalPodAutoscalerReconciler reconciles a CustomHorizontalPodAutoscaler object
type CustomHorizontalPodAutoscalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomHorizontalPodAutoscaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CustomHorizontalPodAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var customHPA customautoscalingv1.CustomHorizontalPodAutoscaler
	err := r.Get(ctx, req.NamespacedName, &customHPA)

	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "unable to get CustomHorizontalPodAutoscaler", "name", req.Namespace)
		return ctrl.Result{}, err
	}

	if !customHPA.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomHorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customautoscalingv1.CustomHorizontalPodAutoscaler{}).
		Complete(r)
}

// reconcileHorizontalPodAutoscaler is a reconcile function for horizontal pod autoscaling
func (r *CustomHorizontalPodAutoscalerReconciler) reconcileHorizontalPodAutoscaler(ctx context.Context, customHPA customautoscalingv1.CustomHorizontalPodAutoscaler) error {
	_ = log.FromContext(ctx)

	return nil
}
