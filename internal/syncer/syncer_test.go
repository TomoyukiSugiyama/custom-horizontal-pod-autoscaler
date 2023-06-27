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

package syncer

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	"sample.com/custom-horizontal-pod-autoscaler/internal/metrics"
	"sample.com/custom-horizontal-pod-autoscaler/test/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Syncer", func() {
	ctx := context.Background()

	It("Should get desiredMinMaxReplicas", func() {
		mt := customautoscalingv1.Condition{
			Type: "training",
			Id:   "7-21",
		}
		persedQueryResults := map[customautoscalingv1.Condition]string{mt: "1"}
		metricsCollector := metrics.FakeNewCollector(persedQueryResults)
		namespacedName := types.NamespacedName{Namespace: "dummy-namespace", Name: "test-customhpa"}
		syncer, err := New(metricsCollector, k8sClient, namespacedName, WithSyncersInterval(10*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())
		go syncer.Start(ctx)
		defer syncer.Stop()
		time.Sleep(20 * time.Millisecond)
		desiredMinMaxremplicas := syncer.GetDesiredMinMaxReplicas()
		Expect(desiredMinMaxremplicas.MinReplicas).Should(Equal(pointer.Int32Ptr(5)))
		Expect(desiredMinMaxremplicas.MaxReplicas).Should(Equal(pointer.Int32(10)))
	})

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &customautoscalingv1.CustomHorizontalPodAutoscaler{}, client.InNamespace("dummy-namespace"))

		customHorizontalPodAutoscaler := util.NewCustomHorizontalPodAutoscaler()
		err = k8sClient.Create(ctx, customHorizontalPodAutoscaler)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
	})
})
