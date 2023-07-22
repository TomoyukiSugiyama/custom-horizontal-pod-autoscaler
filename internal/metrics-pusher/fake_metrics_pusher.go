package pusher

import (
	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
)

type FakeMetricsPusher struct {
}

// Guarantee *FakeMetricsPusher implements MetricsPusher.
var _ MetricsPusher = (*FakeMetricsPusher)(nil)

func FakeNewCollector() *FakeMetricsPusher {
	return &FakeMetricsPusher{}
}

func (p *FakeMetricsPusher) SetSyncerTotal(count float64) {

}

func (p *FakeMetricsPusher) SetCollectorStatus(namespace string, name string, status customautoscalingv1alpha1.CollectorStatus) {

}
