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
	"sync"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	autoscalingv2apply "k8s.io/client-go/applyconfigurations/autoscaling/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	metricspkg "sample.com/custom-horizontal-pod-autoscaler/internal/metrics"
)

// CustomHorizontalPodAutoscalerReconciler reconciles a CustomHorizontalPodAutoscaler object
type CustomHorizontalPodAutoscalerReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	metricsCollector  metricspkg.MetricsCollector
	metricsJobClients map[types.NamespacedName]metricspkg.MetricsJobClient

	mu sync.RWMutex
}

type Option func(*CustomHorizontalPodAutoscalerReconciler)

//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=custom-autoscaling.sample.com,resources=customhorizontalpodautoscalers/finalizers,verbs=update

func NewReconcile(Client client.Client, Scheme *runtime.Scheme, metricsCollector metricspkg.MetricsCollector, opts ...Option) *CustomHorizontalPodAutoscalerReconciler {

	r := &CustomHorizontalPodAutoscalerReconciler{
		Client:            Client,
		Scheme:            Scheme,
		metricsCollector:  metricsCollector,
		metricsJobClients: make(map[types.NamespacedName]metricspkg.MetricsJobClient),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithMetricsJobClients(metricsJobClients map[types.NamespacedName]metricspkg.MetricsJobClient) Option {
	return func(r *CustomHorizontalPodAutoscalerReconciler) {
		r.metricsJobClients = metricsJobClients
	}
}

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
	r.mu.RLock()
	metricsJobClient, metricsJobClientExists := r.metricsJobClients[req.NamespacedName]
	r.mu.RUnlock()

	var customHPA customautoscalingv1.CustomHorizontalPodAutoscaler
	err := r.Get(ctx, req.NamespacedName, &customHPA)
	if errors.IsNotFound(err) {
		if metricsJobClientExists {
			metricsJobClient.Stop()
			r.mu.Lock()
			delete(r.metricsJobClients, req.NamespacedName)
			r.mu.Unlock()
		}
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "unable to get CustomHorizontalPodAutoscaler")
		return ctrl.Result{}, err
	}

	// TODO: Need to check validation of customHPA

	if !customHPA.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if !metricsJobClientExists {
		// TODO: Need to set interval from main.
		metricsJobClient, err = metricspkg.New(
			r.metricsCollector,
			metricspkg.WithInterval(30*time.Second),
			metricspkg.WithCustomHPA(customHPA),
			metricspkg.WithCtrlClient(r.Client),
		)
		if err != nil {
			return ctrl.Result{}, err
		}
		go metricsJobClient.Start(ctx)
		r.mu.Lock()
		r.metricsJobClients[req.NamespacedName] = metricsJobClient
		r.mu.Unlock()
		logger.Info("create metricsJobClient successfully")
	}

	err = r.reconcileHorizontalPodAutoscaler(ctx, customHPA, metricsJobClient)
	if err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, customHPA)
}

// reconcileHorizontalPodAutoscaler is a reconcile function for horizontal pod autoscaling
func (r *CustomHorizontalPodAutoscalerReconciler) reconcileHorizontalPodAutoscaler(
	ctx context.Context,
	customHPA customautoscalingv1.CustomHorizontalPodAutoscaler,
	metricsJobClient metricspkg.MetricsJobClient,
) error {
	// TODO: reconcile hpa if hpa is deleted.
	logger := log.FromContext(ctx)
	hpaName := customHPA.Spec.HorizontalPodAutoscalerName

	desiredMinMaxReplicas := metricsJobClient.GetDesiredMinMaxReplicas()
	minReplicas := customHPA.Spec.MinReplicas
	if desiredMinMaxReplicas.MinReplicas != nil {
		minReplicas = desiredMinMaxReplicas.MinReplicas
	}

	maxReplicas := customHPA.Spec.MaxReplicas
	// TODO: change MaxReplicas type to *int32
	if desiredMinMaxReplicas.MaxReplicas != 0 {
		maxReplicas = desiredMinMaxReplicas.MaxReplicas
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: autoscalingv2.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hpaName,
			Namespace: customHPA.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&customHPA, customautoscalingv1.SchemeBuilder.GroupVersion.WithKind("CustomHorizontalPodAutoscaler")),
			},
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas:    minReplicas,
			MaxReplicas:    maxReplicas,
			ScaleTargetRef: customHPA.Spec.ScaleTargetRef,
			Metrics:        customHPA.Spec.Metrics,
			Behavior:       customHPA.Spec.Behavior.DeepCopy(),
		},
	}

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

	logger.Info("reconcile HorizontalPodAutoscaler successfully", "name", hpaName)
	return nil
}

func (r *CustomHorizontalPodAutoscalerReconciler) updateStatus(
	ctx context.Context,
	customHPA customautoscalingv1.CustomHorizontalPodAutoscaler,
) (ctrl.Result, error) {
	var current autoscalingv2.HorizontalPodAutoscaler
	err := r.Get(ctx, client.ObjectKey{Namespace: customHPA.Namespace, Name: customHPA.Spec.HorizontalPodAutoscalerName}, &current)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Need to set initial value to CurrentReplicas
	status := customautoscalingv1.CustomHorizontalPodAutoscalerStatus{
		CurrentReplicas:    current.Status.CurrentReplicas,
		DesiredReplicas:    current.Status.DesiredReplicas,
		CurrentMinReplicas: *current.Spec.MinReplicas,
		DesiredMinReplicas: customHPA.Status.DesiredMinReplicas,
		CueerntMaxReplicas: current.Spec.MaxReplicas,
		DesiredMaxReplicas: customHPA.Status.DesiredMaxReplicas,
		LastScaleTime:      current.Status.LastScaleTime,
		CurrentMetrics:     current.Status.CurrentMetrics,
		Conditions:         current.Status.Conditions,
		ObservedGeneration: current.Status.ObservedGeneration,
	}

	customHPA.Status = status
	err = r.Status().Update(ctx, &customHPA)
	if err != nil {
		return ctrl.Result{}, err
	}

	if customHPA.Spec.MinReplicas != current.Spec.MinReplicas {
		return ctrl.Result{Requeue: true}, nil
	}

	if customHPA.Spec.MaxReplicas != current.Spec.MaxReplicas {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomHorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customautoscalingv1.CustomHorizontalPodAutoscaler{}).
		Complete(r)
}
