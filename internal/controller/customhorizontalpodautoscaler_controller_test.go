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
	"time"

	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	metricspkg "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/internal/metrics"
	syncerpkg "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/internal/syncer"
	"github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CustomHorizontalPodAutoscaler controller", func() {
	ctx := context.Background()
	var stopFunc func()

	It("Should create HorizontalPodAutoscaler", func() {
		customHorizontalPodAutoscaler := util.NewCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(100 * time.Millisecond)
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

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1alpha1.CustomHorizontalPodAutoscaler{}, client.InNamespace("dummy-namespace"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme.Scheme,
			MetricsBindAddress: "0",
		})
		Expect(err).NotTo(HaveOccurred())

		desiredSpec := customautoscalingv1alpha1.ConditionalReplicasSpec{
			MinReplicas: pointer.Int32(1),
			MaxReplicas: pointer.Int32(5),
		}
		fakeSyncer := syncerpkg.FakeNew(desiredSpec)
		namespacedName := types.NamespacedName{Namespace: "dummy-namespace", Name: "test-customhpa"}
		fakeSyncers := map[types.NamespacedName]syncerpkg.Syncer{namespacedName: fakeSyncer}

		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: "http://localhost:9090"})
		Expect(err).NotTo(HaveOccurred())

		api := prometheusv1.NewAPI(client)
		dummyCollector, err := metricspkg.NewCollector(api)
		Expect(err).NotTo(HaveOccurred())

		reconciler := NewReconcile(k8sClient, scheme.Scheme, dummyCollector, WithSyncers(fakeSyncers))

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
		time.Sleep(100 * time.Millisecond)
	})
})
