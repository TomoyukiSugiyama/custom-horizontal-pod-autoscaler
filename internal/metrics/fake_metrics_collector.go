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

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

type FakeMetricsCollector struct {
	stopCh             chan struct{}
	persedQueryResults map[customautoscalingv1.Condition]string
	mu                 sync.RWMutex
}

// Guarantee *FakeMetricsCollector implements MetricsCollector.
var _ MetricsCollector = (*FakeMetricsCollector)(nil)

func FakeNewCollector(persedQueryResults map[customautoscalingv1.Condition]string) *FakeMetricsCollector {
	return &FakeMetricsCollector{
		persedQueryResults: persedQueryResults,
	}
}

func (c *FakeMetricsCollector) GetPersedQueryResult() map[customautoscalingv1.Condition]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.persedQueryResults
}

func (c *FakeMetricsCollector) Start(ctx context.Context) {
}

func (c *FakeMetricsCollector) Stop() {
}
