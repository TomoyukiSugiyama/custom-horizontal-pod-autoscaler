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

	ctrl "sigs.k8s.io/controller-runtime"

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	jobpkg "sample.com/custom-horizontal-pod-autoscaler/internal/job"
)

var _ = Describe("CustomHorizontalPodAutoscaler controller", func() {
	ctx := context.Background()
	var stopFunc func()

	It("Should create HorizontalPodAutoscaler", func() {
		customHorizontalPodAutoscaler := newCustomHorizontalPodAutoscaler()
		err := k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		fakeJobClient := jobpkg.FakeNew(k8sClient, "dummy-namespace", "sample")
		fakeJobClient.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
		hpa := autoscalingv2.HorizontalPodAutoscaler{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "dummy-namespace", Name: "sample"}, &hpa)
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

		fakeJobClient := jobpkg.FakeNew(k8sClient, "dummy-namespace", "sample")
		namespacedName := types.NamespacedName{Namespace: "dummy-namespace", Name: "sample"}
		fakeJobClients := map[types.NamespacedName]jobpkg.JobClient{namespacedName: fakeJobClient}

		reconciler := NewReconcile(k8sClient, scheme.Scheme, WithJobClients(fakeJobClients))

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
	minReplicasTraining := int32(7)
	maxReplicas := int32(5)
	maxReplicasTraining := int32(10)
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
			MinReplicas:           &minReplicas,
			MinReplicasTraining:   &minReplicasTraining,
			MaxReplicas:           maxReplicas,
			MaxReplicasTraining:   maxReplicasTraining,
			ScaleTargetRef:        scaleTargetRef,
			Metrics:               metrics,
			TemporaryScaleMetrics: temporaryScaleMetrics,
		},
	}
}
