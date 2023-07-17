package pusher

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type MetricsPusher interface {
	SetSyncerTotal()
}

type metricsPusher struct {
	syncerTotal *prometheus.GaugeVec
}

type PusherOption func(*metricsPusher)

func NewPusher(opts ...PusherOption) (MetricsPusher, error) {
	p := &metricsPusher{}

	p.syncerTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "syncertotal",
		Help:      "Syncer Count",
	}, []string{"name", "namespace"})
	metrics.Registry.MustRegister(p.syncerTotal)

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

func (p *metricsPusher) SetSyncerTotal() {
	p.syncerTotal.WithLabelValues("test", "testns").Set(1)
}
