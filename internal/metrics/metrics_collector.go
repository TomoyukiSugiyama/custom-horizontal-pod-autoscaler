package metrics

import (
	"context"
	"time"

	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MetricsCollector interface {
	Start(ctx context.Context)
	Stop()
	GetPersedQueryResult() map[metricType]string
}

type metricsCollector struct {
	prometheusApi      prometheusv1.API
	interval           time.Duration
	stopCh             chan struct{}
	query              string
	persedQueryResults map[metricType]string
}

type metricType struct {
	duration string
	jobType  string
}

type CollectorOption func(*metricsCollector)

func NewCollector(prometheusApi prometheusv1.API, opts ...CollectorOption) (MetricsCollector, error) {
	c := &metricsCollector{
		prometheusApi: prometheusApi,
		interval:      30 * time.Second,
		stopCh:        make(chan struct{}),
		query:         "temporary_scale",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func WithControllerInterval(interval time.Duration) CollectorOption {
	return func(c *metricsCollector) {
		c.interval = interval
	}
}

func (c *metricsCollector) GetPersedQueryResult() map[metricType]string {
	return c.persedQueryResults
}

func (c *metricsCollector) getTemporaryScaleMetrics(ctx context.Context) {
	logger := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	queryResult, warning, err := c.prometheusApi.Query(ctx, c.query, time.Now())
	if err != nil {
		logger.Error(err, "unable to get temporary scale metrics")
		return
	}
	if len(warning) > 0 {
		logger.Info("get wornings", "worning", warning)
	}

	// ref: https://github.com/prometheus/client_golang/issues/1011
	c.perseMetrics(queryResult.(model.Vector))
	for key, queryResult := range c.persedQueryResults {
		logger.Info(
			"parsed query result",
			"duration", key.duration,
			"type", key.jobType,
			"value", queryResult,
		)
	}
}

func (c *metricsCollector) perseMetrics(samples model.Vector) error {
	c.persedQueryResults = make(map[metricType]string)
	for _, sample := range samples {
		metrics, err := parser.ParseMetric(sample.Metric.String())
		if err != nil {
			return err
		}
		k := metricType{
			jobType:  metrics.Map()["type"],
			duration: metrics.Map()["duration"],
		}
		c.persedQueryResults[k] = sample.Value.String()
	}
	return nil
}

func (c *metricsCollector) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting metricsCollector")
	defer logger.Info("shut down metricsCollector")

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("start get metrics ticker")
			c.getTemporaryScaleMetrics(ctx)
		case <-c.stopCh:
			logger.Info("stop metrics collector")
			return
		}
	}
}

func (c *metricsCollector) Stop() {
	close(c.stopCh)
}
