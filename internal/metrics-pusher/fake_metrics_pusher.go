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

package pusher

import (
	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

type FakeMetricsPusher struct {
}

// Guarantee *FakeMetricsPusher implements MetricsPusher.
var _ MetricsPusher = (*FakeMetricsPusher)(nil)

func FakeNewPusher() *FakeMetricsPusher {
	return &FakeMetricsPusher{}
}

func (p *FakeMetricsPusher) SetSyncerTotal(count float64) {

}

func (p *FakeMetricsPusher) SetCollectorStatus(status customautoscalingv1alpha1.CollectorStatus) {

}

func (p *FakeMetricsPusher) GetCollectorNotReady() *prometheus.GaugeVec {
	return nil
}

func (p *FakeMetricsPusher) GetCollectorAvailable() *prometheus.GaugeVec {
	return nil
}

func (p *FakeMetricsPusher) GetSyncerTotal() *prometheus.GaugeVec {
	return nil
}
