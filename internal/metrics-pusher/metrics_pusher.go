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
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type MetricsPusher interface {
	SetSyncerTotal(count float64)
	SetCollectorStatus(status customautoscalingv1alpha1.CollectorStatus)
	GetCollectorNotReady() *prometheus.GaugeVec
	GetCollectorAvailable() *prometheus.GaugeVec
	GetSyncerTotal() *prometheus.GaugeVec
}

type metricsPusher struct {
	syncerTotal        *prometheus.GaugeVec
	collectorNotReady  *prometheus.GaugeVec
	collectorAvailable *prometheus.GaugeVec
}

func NewPusher() (MetricsPusher, error) {
	p := &metricsPusher{}

	p.syncerTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "syncer_total",
		Help:      "Total number of syncers per controller",
	}, []string{"controller"})

	p.collectorNotReady = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "collector_notready",
		Help:      "The controller status about not ready condition",
	}, []string{"controller"})

	p.collectorAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "customhpa",
		Name:      "collector_available",
		Help:      "The controller status about available condition",
	}, []string{"controller"})

	metrics.Registry.MustRegister(p.syncerTotal, p.collectorNotReady, p.collectorAvailable)

	return p, nil
}

func (p *metricsPusher) SetSyncerTotal(count float64) {
	p.syncerTotal.WithLabelValues("customhorizontalpodautoscaler").Set(count)
}

func (p *metricsPusher) SetCollectorStatus(status customautoscalingv1alpha1.CollectorStatus) {
	switch status {
	case customautoscalingv1alpha1.CollectorNotReady:
		p.collectorNotReady.WithLabelValues("customhorizontalpodautoscaler").Set(1)
		p.collectorAvailable.WithLabelValues("customhorizontalpodautoscaler").Set(0)
	case customautoscalingv1alpha1.CollectorAvailable:
		p.collectorNotReady.WithLabelValues("customhorizontalpodautoscaler").Set(0)
		p.collectorAvailable.WithLabelValues("customhorizontalpodautoscaler").Set(1)
	}
}

func (p *metricsPusher) GetCollectorNotReady() *prometheus.GaugeVec {
	return p.collectorNotReady
}

func (p *metricsPusher) GetCollectorAvailable() *prometheus.GaugeVec {
	return p.collectorAvailable
}

func (p *metricsPusher) GetSyncerTotal() *prometheus.GaugeVec {
	return p.syncerTotal
}
