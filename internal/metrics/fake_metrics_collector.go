package metrics

import (
	"context"
	"sync"
)

type FakeMetricsCollector struct {
	stopCh             chan struct{}
	persedQueryResults map[metricType]string
	mu                 sync.RWMutex
}

// Guarantee *FakeMetricsCollector implements MetricsCollector.
var _ MetricsCollector = (*FakeMetricsCollector)(nil)

func FakeNewCollector(persedQueryResults map[metricType]string) *FakeMetricsCollector {
	return &FakeMetricsCollector{
		persedQueryResults: persedQueryResults,
	}
}

func (c *FakeMetricsCollector) GetPersedQueryResult() map[metricType]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.persedQueryResults
}

func (c *FakeMetricsCollector) Start(ctx context.Context) {
}

func (c *FakeMetricsCollector) Stop() {
}
