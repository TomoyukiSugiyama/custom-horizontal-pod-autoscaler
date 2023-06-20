package job

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type JobClient interface {
	Start(ctx context.Context)
	Stop()
}

type jobClient struct {
	api      v1.API
	interval time.Duration
	stopCh   chan struct{}
	ts       temporaryScale
}

type temporaryScale struct {
	duration string
	jobType  string
	value    string
}

type Option func(*jobClient)

func New(opts ...Option) (JobClient, error) {

	// TODO: Need to set api address from main.
	client, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	if err != nil {
		return nil, err
	}
	api := v1.NewAPI(client)

	j := &jobClient{
		api:      api,
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(j)
	}

	return j, nil
}

func WithInterval(interval time.Duration) Option {
	return func(j *jobClient) {
		j.interval = interval
	}
}

func (j *jobClient) getTemporaryScaleMetrics(ctx context.Context) {
	logger := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	query := "temporary_scale"
	queryResult, warning, err := j.api.Query(ctx, query, time.Now())
	if err != nil {
		logger.Error(err, "unable to get temporary scale metrics")
		return
	}
	if len(warning) > 0 {
		logger.Info("get wornings", "worning", warning)
	}

	// ref: https://github.com/prometheus/client_golang/issues/1011
	j.perseMetrics(queryResult.(model.Vector))
	logger.Info("get metrics parse", "duration", j.ts.duration, "type", j.ts.jobType, "value", j.ts.value)
}

func (j *jobClient) perseMetrics(samples model.Vector) error {
	if len(samples) != 1 {
		return errors.New("multiple sample")
	}
	j.ts.value = samples[0].Value.String()
	metrics, err := parser.ParseMetric(samples[0].Metric.String())
	if err != nil {
		return err
	}
	j.ts.duration = metrics.Map()["duration"]
	j.ts.jobType = metrics.Map()["type"]
	return nil
}

func (j *jobClient) Start(ctx context.Context) {
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
		case <-j.stopCh:
			logger.Info("received stop signal")
			return
		}
	}
}

func (c *jobClient) Stop() {
	close(c.stopCh)
}
