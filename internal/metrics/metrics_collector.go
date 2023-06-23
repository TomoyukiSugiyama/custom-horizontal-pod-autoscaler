package metrics

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MetricsCollector interface {
	Start(ctx context.Context)
	Stop()
}

type metricsCollector struct {
	interval time.Duration
	stopCh   chan struct{}
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
			logger.Info("collect metrics")
		case <-c.stopCh:
			logger.Info("stop metrics collector")
			return
		}
	}
}

func (c *metricsCollector) Stop() {
	close(c.stopCh)
}
