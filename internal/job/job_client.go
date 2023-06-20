package job

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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
}

type Option func(*jobClient)

func New(opts ...Option) (JobClient, error) {

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
	now := time.Now()
	rangeParam := v1.Range{Start: now.Add(-time.Hour), End: now, Step: j.interval}
	_, warning, err := j.api.QueryRange(ctx, query, rangeParam)
	if err != nil {
		logger.Error(err, "unable to get temporary scale metrics")
		return
	}
	if len(warning) > 0 {
		logger.Info("get wornings", "worning", warning)
	}
	logger.Info("get metrics")
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
