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

func (p *FakeMetricsPusher) SetCollectorStatus(namespace string, name string, status customautoscalingv1alpha1.CollectorStatus) {

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
