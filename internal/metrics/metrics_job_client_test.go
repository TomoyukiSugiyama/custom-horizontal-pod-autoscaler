package metrics

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MetricsJobClient", func() {
	ctx := context.Background()
	//var stopFunc func()

	It("Should get desiredMinMaxremplicas", func() {
		mt := metricType{
			jobType:  "training",
			duration: "7-21",
		}
		persedQueryResults := map[metricType]string{mt: "1"}
		metricsCollector := FakeNewCollector(persedQueryResults)
		namespacedName := types.NamespacedName{Namespace: "dummy-namespace", Name: "test-customhpa"}
		metricsJobClient, err := New(metricsCollector, k8sClient, namespacedName, WithMetricsJobClientsInterval(10*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())
		go metricsJobClient.Start(ctx)
		time.Sleep(20 * time.Millisecond)
		desiredMinMaxremplicas := metricsJobClient.GetDesiredMinMaxReplicas()
		Expect(desiredMinMaxremplicas.MinReplicas).Should(Equal(pointer.Int32Ptr(5)))
		Expect(desiredMinMaxremplicas.MaxReplicas).Should(Equal(pointer.Int32(10)))
		metricsJobClient.Stop()
	})

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1.CustomHorizontalPodAutoscaler{}, client.InNamespace("dummy-namespace"))

		customHorizontalPodAutoscaler := newCustomHorizontalPodAutoscaler()
		err = k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		//stopFunc()
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
			MaxReplicas: &workdayMaxRelpicas,
		},
		{
			Type:        "training",
			Duration:    "7-21",
			MinReplicas: &trainingMinRelpicas,
			MaxReplicas: &trainingMaxRelpicas,
		},
	}

	return &customautoscalingv1.CustomHorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-customhpa",
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
