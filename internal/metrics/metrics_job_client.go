package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
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
	ctrlClient                      ctrlClient.Client
	api                             v1.API
	interval                        time.Duration
	stopCh                          chan struct{}
	query                           string
	persedQueryResults              map[metricType]string
	desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec
	customHPA                       customautoscalingv1.CustomHorizontalPodAutoscaler
}

var _ MetricsJobClient = (*metricsJobClient)(nil)

type metricType struct {
	duration string
	jobType  string
}

type Option func(*metricsJobClient)

func New(opts ...Option) (MetricsJobClient, error) {

	// TODO: Need to set api address from main.
	client, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	if err != nil {
		return nil, err
	}
	api := v1.NewAPI(client)

	j := &metricsJobClient{
		api:                             api,
		interval:                        30 * time.Second,
		stopCh:                          make(chan struct{}),
		query:                           "temporary_scale",
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

func WithQuery(query string) Option {
	return func(j *metricsJobClient) {
		j.query = query
	}
}

func WithCtrlClient(ctrlClient ctrlClient.Client) Option {
	return func(j *metricsJobClient) {
		j.ctrlClient = ctrlClient
	}
}

func WithCustomHPA(customHPA customautoscalingv1.CustomHorizontalPodAutoscaler) Option {
	return func(j *metricsJobClient) {
		j.customHPA = customHPA
	}
}

func (j *metricsJobClient) getTemporaryScaleMetrics(ctx context.Context) {
	logger := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	queryResult, warning, err := j.api.Query(ctx, j.query, time.Now())
	if err != nil {
		logger.Error(err, "unable to get temporary scale metrics")
		return
	}
	if len(warning) > 0 {
		logger.Info("get wornings", "worning", warning)
	}

	// ref: https://github.com/prometheus/client_golang/issues/1011
	j.perseMetrics(queryResult.(model.Vector))
	j.updateDesiredMinMaxReplicas()
	for key, queryResult := range j.persedQueryResults {
		logger.Info(
			"parsed query result",
			"duration", key.duration,
			"type", key.jobType,
			"value", queryResult,
		)
	}

	for _, metric := range j.customHPA.Spec.TemporaryScaleMetrics {
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

func (j *metricsJobClient) perseMetrics(samples model.Vector) error {
	j.persedQueryResults = make(map[metricType]string)
	for _, sample := range samples {
		metrics, err := parser.ParseMetric(sample.Metric.String())
		if err != nil {
			return err
		}
		k := metricType{
			jobType:  metrics.Map()["type"],
			duration: metrics.Map()["duration"],
		}
		j.persedQueryResults[k] = sample.Value.String()
	}
	return nil
}

func (j *metricsJobClient) updateDesiredMinMaxReplicas() {
	for _, m := range j.customHPA.Spec.TemporaryScaleMetrics {
		k := metricType{
			jobType:  m.Type,
			duration: m.Duration,
		}
		v, isExist := j.persedQueryResults[k]
		if isExist && v == "1" {
			j.desiredTemporaryScaleMetricSpec.MinReplicas = m.MinReplicas
			j.desiredTemporaryScaleMetricSpec.MaxReplicas = m.MaxReplicas
			return
		}
	}
	j.desiredTemporaryScaleMetricSpec.MinReplicas = j.customHPA.Spec.MinReplicas
	j.desiredTemporaryScaleMetricSpec.MaxReplicas = j.customHPA.Spec.MaxReplicas
}

func (j *metricsJobClient) updateStatus(ctx context.Context) error {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, ctrlClient.ObjectKey{Namespace: j.customHPA.Namespace, Name: j.customHPA.Name}, &current)
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
