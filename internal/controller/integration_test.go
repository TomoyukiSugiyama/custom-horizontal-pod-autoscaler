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

package controller

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	collectorpkg "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/internal/metrics-collector"
	pusherpkg "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/internal/metrics-pusher"
	"github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("Integration test", func() {
	ctx := context.Background()
	var stopFunc func()
	var fakePrometheus *httptest.Server
	var collector collectorpkg.MetricsCollector
	pusher, _ := pusherpkg.NewPusher()
	metricPort := 13449

	It("Should create HorizontalPodAutoscaler", func() {
		customHorizontalPodAutoscaler := util.NewCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(20 * time.Millisecond)
		hpa := autoscalingv2.HorizontalPodAutoscaler{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-hpa"}, &hpa)
		}).Should(Succeed())

		Expect(hpa.Name).Should(Equal("test-hpa"))
		Expect(hpa.Spec.MinReplicas).Should(Equal(pointer.Int32Ptr(1)))
		Expect(hpa.Spec.MaxReplicas).Should(Equal(int32(5)))
		expectedScaleTargetRef := autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "test-deployment",
		}
		Expect(hpa.Spec.ScaleTargetRef).Should(Equal(expectedScaleTargetRef))
		expectedMetrics := []autoscalingv2.MetricSpec{
			{
				Type: "Pods",
				Pods: &autoscalingv2.PodsMetricSource{
					Metric: autoscalingv2.MetricIdentifier{
						Name: "memory_usage_bytes",
					},
					Target: autoscalingv2.MetricTarget{
						Type:         "AverageValue",
						AverageValue: resource.NewQuantity(8*1024*1024, resource.BinarySI),
					},
				},
			},
		}
		Expect(hpa.Spec.Metrics).Should(Equal(expectedMetrics))
	})

	It("Should exist HorizontalPodAutoscaler after delete", func() {
		customHorizontalPodAutoscaler := util.NewCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
		hpa := autoscalingv2.HorizontalPodAutoscaler{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-hpa"}, &hpa)
		}).Should(Succeed())
		time.Sleep(100 * time.Millisecond)
		Eventually(func() error {
			return k8sClient.Delete(ctx, &hpa)
		}).Should(Succeed())
		time.Sleep(100 * time.Millisecond)
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-hpa"}, &hpa)
		}).Should(Succeed())

		Expect(hpa.Name).Should(Equal("test-hpa"))
	})

	It("Should not exist CustomHorizontalPodAutoscaler after delete", func() {
		customHorizontalPodAutoscaler := util.NewCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
		customHPA := customHorizontalPodAutoscaler
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-customhpa"}, customHPA)
		}).Should(Succeed())
		time.Sleep(100 * time.Millisecond)
		Eventually(func() error {
			return k8sClient.Delete(ctx, customHPA)
		}).Should(Succeed())
		time.Sleep(100 * time.Millisecond)
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-customhpa"}, customHPA)
		}).ShouldNot(Succeed())
	})

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1alpha1.CustomHorizontalPodAutoscaler{}, client.InNamespace("dummy-namespace"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: fmt.Sprintf("localhost:%d", metricPort)},
		})
		metricPort--
		Expect(err).NotTo(HaveOccurred())

		fakePrometheus = util.NewFakePrometheusServer()

		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: fakePrometheus.URL})
		Expect(err).NotTo(HaveOccurred())
		api := prometheusv1.NewAPI(client)
		collector, err = collectorpkg.NewCollector(api, collectorpkg.WithMetricsCollectorInterval(50*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		reconciler := NewReconcile(k8sClient, scheme.Scheme, collector, WithSyncersInterval(50*time.Millisecond), WithMetricsPusher(pusher))

		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

	})

	AfterEach(func() {
		stopFunc()
		collector.Stop()
		fakePrometheus.Close()
		time.Sleep(100 * time.Millisecond)
	})
})
