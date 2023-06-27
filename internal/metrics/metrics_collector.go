/*
Copyright 2023 Tomoyuki Sugiyama.

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

package metrics

import (
	"context"
	"sync"
	"time"

	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MetricsCollector interface {
	Start(ctx context.Context)
	Stop()
	GetPersedQueryResult() map[MetricType]string
}

type metricsCollector struct {
	prometheusApi      prometheusv1.API
	interval           time.Duration
	stopCh             chan struct{}
	query              string
	persedQueryResults map[MetricType]string
	mu                 sync.RWMutex
}

type MetricType struct {
	Id      string
	JobType string
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

func WithMetricsCollectorInterval(interval time.Duration) CollectorOption {
	return func(c *metricsCollector) {
		c.interval = interval
	}
}

func (c *metricsCollector) GetPersedQueryResult() map[MetricType]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
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
			"id", key.Id,
			"type", key.JobType,
			"value", queryResult,
		)
	}
}

func (c *metricsCollector) perseMetrics(samples model.Vector) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.persedQueryResults = make(map[MetricType]string)
	for _, sample := range samples {
		metrics, err := parser.ParseMetric(sample.Metric.String())
		if err != nil {
			return err
		}
		k := MetricType{
			JobType: metrics.Map()["type"],
			Id:      metrics.Map()["id"],
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
