package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"

	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	metricspkg "sample.com/custom-horizontal-pod-autoscaler/internal/metrics"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("CustomHorizontalPodAutoscaler controller", func() {
	ctx := context.Background()
	var stopFunc func()

	It("Should create HorizontalPodAutoscaler", func() {
		customHorizontalPodAutoscaler := newCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())
		hpa := autoscalingv2.HorizontalPodAutoscaler{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "test-hpa"}, &hpa)
		}).Should(Succeed())
		Expect(hpa.Spec.MinReplicas).Should(Equal(pointer.Int32Ptr(1)))
		Expect(hpa.Spec.MaxReplicas).Should(Equal(int32(5)))
	})

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1.CustomHorizontalPodAutoscaler{}, client.InNamespace("dummy-namespace"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).NotTo(HaveOccurred())

		desiredSpec := apiv1.TemporaryScaleMetricSpec{
			MinReplicas: pointer.Int32(1),
			MaxReplicas: int32(5),
		}
		fakeMetricsJobClient := metricspkg.FakeNew(desiredSpec)
		namespacedName := types.NamespacedName{Namespace: "dummy-namespace", Name: "sample"}
		fakeMetricsJobClients := map[types.NamespacedName]metricspkg.MetricsJobClient{namespacedName: fakeMetricsJobClient}

		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: "http://localhost:9090"})
		Expect(err).NotTo(HaveOccurred())

		api := prometheusv1.NewAPI(client)
		collector, err := metricspkg.NewCollector(api)
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)

		reconciler := NewReconcile(k8sClient, scheme.Scheme, collector, WithMetricsJobClients(fakeMetricsJobClients))

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

func newCustomHorizontalPodAutoscaler() *customautoscalingv1.CustomHorizontalPodAutoscaler {
	minReplicas := int32(1)
	maxReplicas := int32(5)
	scaleTargetRef := autoscalingv2.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "test-deployment",
	}
	metrics := []autoscalingv2.MetricSpec{
		{
			Type: "Pods",
			Pods: &autoscalingv2.PodsMetricSource{
				Metric: autoscalingv2.MetricIdentifier{
					Name: "memory_usage_bytes",
				},
				Target: autoscalingv2.MetricTarget{
					Type:         "Value",
					AverageValue: resource.NewQuantity(8*1024*1024, resource.BinarySI),
				},
			},
		},
	}
	workdayMinRelpicas := int32(2)
	workdayMaxRelpicas := int32(4)
	trainingMinRelpicas := int32(5)
	trainingMaxRelpicas := int32(10)
	temporaryScaleMetrics := []customautoscalingv1.TemporaryScaleMetricSpec{
		{
			Type:        "workday",
			Duration:    "7-21",
			MinReplicas: &workdayMinRelpicas,
			MaxReplicas: workdayMaxRelpicas,
		},
		{
			Type:        "training",
			Duration:    "7-21",
			MinReplicas: &trainingMinRelpicas,
			MaxReplicas: trainingMaxRelpicas,
		},
	}

	return &customautoscalingv1.CustomHorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "dummy-namespace",
		},
		Spec: customautoscalingv1.CustomHorizontalPodAutoscalerSpec{
			HorizontalPodAutoscalerName: "test-hpa",
			MinReplicas:                 &minReplicas,
			MaxReplicas:                 maxReplicas,
			ScaleTargetRef:              scaleTargetRef,
			Metrics:                     metrics,
			TemporaryScaleMetrics:       temporaryScaleMetrics,
		},
	}
}
