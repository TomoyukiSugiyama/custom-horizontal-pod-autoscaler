package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

var _ = Describe("CustomHorizontalPodAutoscaler controller", func() {
	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1.CustomHorizontalPodAutoscaler{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())

		reconciler := CustomHorizontalPodAutoscalerReconciler{
			Client: k8sClient,
			Scheme: scheme.Scheme,
		}

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
