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

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	autoscalingv2apply "k8s.io/client-go/applyconfigurations/autoscaling/v2"
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

	err = r.reconcileHorizontalPodAutoscaler(ctx, customHPA)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, customHPA)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomHorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customautoscalingv1.CustomHorizontalPodAutoscaler{}).
		Complete(r)
}

// reconcileHorizontalPodAutoscaler is a reconcile function for horizontal pod autoscaling
func (r *CustomHorizontalPodAutoscalerReconciler) reconcileHorizontalPodAutoscaler(ctx context.Context, customHPA customautoscalingv1.CustomHorizontalPodAutoscaler) error {
	logger := log.FromContext(ctx)
	hpaName := customHPA.Name

	hpa := autoscalingv2.HorizontalPodAutoscaler{

		ObjectMeta: metav1.ObjectMeta{
			Name:      customHPA.Name,
			Namespace: customHPA.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&customHPA, customautoscalingv1.SchemeBuilder.GroupVersion.WithKind("CustomHorizontalPodAutoscaler")),
			},
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas:    customHPA.Spec.MinReplicas,
			MaxReplicas:    customHPA.Spec.MaxReplicas,
			ScaleTargetRef: customHPA.Spec.ScaleTargetRef,
			Metrics:        customHPA.Spec.Metrics,
			Behavior:       customHPA.Spec.Behavior.DeepCopy(),
		},
	}

	// hpa := autoscalingv2apply.HorizontalPodAutoscaler(hpaName, customHPA.Namespace).
	// 	WithLabels(map[string]string{
	// 		"app.kubernetes.io/name":       hpaName,
	// 		"app.kubernetes.io/instance":   customHPA.Name,
	// 		"app.kubernetes.io/created-by": "custom-horizontal-pod-autoscaler-controller",
	// 	}).
	// 	WithSpec(autoscalingv2apply.HorizontalPodAutoscalerSpec().
	// 		WithMinReplicas(*customHPA.Spec.MinReplicas).
	// 		WithMaxReplicas(customHPA.Spec.MaxReplicas).
	// 		WithBehavior(
	// 			autoscalingv2apply.HorizontalPodAutoscalerBehavior().
	// 				WithScaleDown(autoscalingv2apply.HPAScalingRules()).
	// 				WithScaleUp(autoscalingv2apply.HPAScalingRules()),
	// 		).
	// 		WithMetrics(autoscalingv2apply.MetricSpec().
	// 			WithContainerResource(autoscalingv2apply.ContainerResourceMetricSource().
	// 				WithContainer(autoscalingv2.SchemeGroupVersion.Group).
	// 				WithName(*autoscalingv2apply.ContainerResourceMetricSource().Name).
	// 				WithTarget(autoscalingv2apply.MetricTarget().WithAverageUtilization(*customHPA.Spec.Metrics[0].ContainerResource.Target.AverageUtilization)),
	// 			).
	// 			WithExternal(autoscalingv2apply.ExternalMetricSource()).
	// 			WithObject(autoscalingv2apply.ObjectMetricSource()).
	// 			WithPods(autoscalingv2apply.PodsMetricSource()).
	// 			WithResource(autoscalingv2apply.ResourceMetricSource()).
	// 			WithType(autoscalingv2.ContainerResourceMetricSourceType),
	// 		).
	// 		WithScaleTargetRef(autoscalingv2apply.CrossVersionObjectReference().
	// 			WithAPIVersion(customHPA.Spec.ScaleTargetRef.APIVersion).
	// 			WithKind(customHPA.Spec.ScaleTargetRef.Kind).
	// 			WithName(customHPA.Spec.ScaleTargetRef.Name),
	// 		),
	// 	)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(hpa)
	if err != nil {
		return err
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current autoscalingv2.HorizontalPodAutoscaler
	err = r.Get(ctx, client.ObjectKey{Namespace: customHPA.Namespace, Name: hpaName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currentApplyConfig, err := autoscalingv2apply.ExtractHorizontalPodAutoscaler(&current, "custom-horizontal-pod-autoscaler-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(hpa, currentApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "custom-horizontal-pod-autoscaler-controller",
		Force:        pointer.Bool(true),
	})

	if err != nil {
		logger.Error(err, "unable to create or update HorizontalPodAutoscaler")
		return err
	}

	logger.Info("reconcile HorizontalPodAutoscaler successfully", "name", customHPA.Name)
	return nil
}

func (r *CustomHorizontalPodAutoscalerReconciler) updateStatus(ctx context.Context, customHPA customautoscalingv1.CustomHorizontalPodAutoscaler) (ctrl.Result, error) {
	var current autoscalingv2.HorizontalPodAutoscaler
	err := r.Get(ctx, client.ObjectKey{Namespace: customHPA.Namespace, Name: customHPA.Name}, &current)
	if err != nil {
		return ctrl.Result{}, err
	}

	customHPA.Status = customautoscalingv1.CustomHorizontalPodAutoscalerStatus(current.Status)
	err = r.Status().Update(ctx, &customHPA)

	if customHPA.Spec.MinReplicas != current.Spec.MinReplicas {
		return ctrl.Result{Requeue: true}, nil
	}

	if customHPA.Spec.MinReplicas != &current.Spec.MaxReplicas {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}
