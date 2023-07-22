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
	syncerTotal        *prometheus.GaugeVec
	collectorNotReady  *prometheus.GaugeVec
	collectorAvailable *prometheus.GaugeVec
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

	p.collectorNotReady = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "collector_notready",
		Help:      "The controller status about not ready condition",
	}, []string{"controller", "name", "namespace"})
	metrics.Registry.MustRegister(p.collectorNotReady)

	p.collectorAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "collector_notready",
		Help:      "The controller status about available condition",
	}, []string{"controller", "name", "namespace"})
	metrics.Registry.MustRegister(p.collectorAvailable)

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

func (p *metricsPusher) SetSyncerTotal(count float64) {
	p.syncerTotal.WithLabelValues("customhorizontalpodautoscaler").Set(count)
}

func (p *metricsPusher) SetCollectorStatus(namespace string, name string, status customautoscalingv1alpha1.CollectorStatus) {
	switch status {
	case customautoscalingv1alpha1.CollectorNotReady:
		p.collectorNotReady.WithLabelValues("customhorizontalpodautoscaler", name, namespace).Set(1)
		p.collectorAvailable.WithLabelValues("customhorizontalpodautoscaler", name, namespace).Set(0)
	case customautoscalingv1alpha1.CollectorAvailable:
		p.collectorNotReady.WithLabelValues("customhorizontalpodautoscaler", name, namespace).Set(0)
		p.collectorAvailable.WithLabelValues("customhorizontalpodautoscaler", name, namespace).Set(1)
	}
}
