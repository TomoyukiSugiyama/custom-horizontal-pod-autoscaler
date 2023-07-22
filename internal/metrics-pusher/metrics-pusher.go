package pusher

import (
	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type MetricsPusher interface {
	SetSyncerTotal(count float64)
}

type metricsPusher struct {
	syncerTotal     *prometheus.GaugeVec
	collectorStatus customautoscalingv1alpha1.CollectorStatus
}

type PusherOption func(*metricsPusher)

func NewPusher(opts ...PusherOption) (MetricsPusher, error) {
	p := &metricsPusher{}

	p.syncerTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "syncer_total",
		Help:      "Total number of syncers per controller",
	}, []string{"controller"})
	metrics.Registry.MustRegister(p.syncerTotal)

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

func (p *metricsPusher) SetSyncerTotal(count float64) {
	p.syncerTotal.WithLabelValues("customhorizontalpodautoscaler").Set(count)
}