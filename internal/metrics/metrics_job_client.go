package metrics

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"
	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MetricsJobClient interface {
	Start(ctx context.Context)
	Stop()
	GetDesiredMinMaxReplicas() apiv1.TemporaryScaleMetricSpec
}

type metricsJobClient struct {
	metricsCollector                MetricsCollector
	ctrlClient                      ctrlClient.Client
	interval                        time.Duration
	stopCh                          chan struct{}
	desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec
	namespacedName                  types.NamespacedName
}

var _ MetricsJobClient = (*metricsJobClient)(nil)

type Option func(*metricsJobClient)

func New(metricsCollector MetricsCollector, ctrlClient ctrlClient.Client, namespacedName types.NamespacedName, opts ...Option) (MetricsJobClient, error) {
	j := &metricsJobClient{
		interval:                        30 * time.Second,
		stopCh:                          make(chan struct{}),
		metricsCollector:                metricsCollector,
		ctrlClient:                      ctrlClient,
		namespacedName:                  namespacedName,
		desiredTemporaryScaleMetricSpec: apiv1.TemporaryScaleMetricSpec{},
	}

	for _, opt := range opts {
		opt(j)
	}

	return j, nil
}

func WithInterval(interval time.Duration) Option {
	return func(j *metricsJobClient) {
		j.interval = interval
	}
}

func (j *metricsJobClient) getTemporaryScaleMetrics(ctx context.Context) {
	logger := log.FromContext(ctx)
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	j.updateDesiredMinMaxReplicas(ctx)

	for _, metric := range current.Spec.TemporaryScaleMetrics {
		logger.Info(
			"customHPA settings",
			"duration", metric.Duration,
			"type", metric.Type,
			"minReplicas", metric.MinReplicas,
			"maxReplicas", metric.MaxReplicas,
		)
	}

	logger.Info(
		"update desired min and max replicas",
		"minReplicas", j.desiredTemporaryScaleMetricSpec.MinReplicas,
		"maxReplicas", j.desiredTemporaryScaleMetricSpec.MaxReplicas,
	)
}

func (j *metricsJobClient) updateDesiredMinMaxReplicas(ctx context.Context) {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)

	for _, m := range current.Spec.TemporaryScaleMetrics {
		k := metricType{
			jobType:  m.Type,
			duration: m.Duration,
		}
		res := j.metricsCollector.GetPersedQueryResult()
		v, isExist := res[k]
		if isExist && v == "1" {
			j.desiredTemporaryScaleMetricSpec.MinReplicas = m.MinReplicas
			j.desiredTemporaryScaleMetricSpec.MaxReplicas = m.MaxReplicas
			return
		}
	}
	j.desiredTemporaryScaleMetricSpec.MinReplicas = current.Spec.MinReplicas
	j.desiredTemporaryScaleMetricSpec.MaxReplicas = current.Spec.MaxReplicas
}

func (j *metricsJobClient) updateStatus(ctx context.Context) error {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	current.Status.DesiredMinReplicas = *j.desiredTemporaryScaleMetricSpec.MinReplicas
	current.Status.DesiredMaxReplicas = j.desiredTemporaryScaleMetricSpec.MaxReplicas
	return j.ctrlClient.Status().Update(ctx, &current)
}

func (j *metricsJobClient) GetDesiredMinMaxReplicas() apiv1.TemporaryScaleMetricSpec {
	return j.desiredTemporaryScaleMetricSpec
}

func (j *metricsJobClient) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting job")
	defer logger.Info("shut down job")

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("received scheduler tick")
			ctx = context.Background()
			j.getTemporaryScaleMetrics(ctx)
			if err := j.updateStatus(ctx); err != nil {
				logger.Error(err, "unable to update status")
			}
		case <-j.stopCh:
			logger.Info("received stop signal")
			return
		}
	}
}

func (j *metricsJobClient) Stop() {
	close(j.stopCh)
}
