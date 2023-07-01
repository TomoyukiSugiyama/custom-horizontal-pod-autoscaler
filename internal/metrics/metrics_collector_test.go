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

package metrics

import (
	"context"
	"net/http/httptest"
	"time"

	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var _ = Describe("Syncer", func() {
	ctx := context.Background()
	var fakePrometheus *httptest.Server

	It("Should get persedQueryResult", func() {
		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: fakePrometheus.URL})

		Expect(err).NotTo(HaveOccurred())
		api := prometheusv1.NewAPI(client)
		collector, err := NewCollector(api, WithMetricsCollectorInterval(50*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)
		defer collector.Stop()
		time.Sleep(100 * time.Millisecond)

		res := collector.GetPersedQueryResult()
		Expect(res[customautoscalingv1alpha1.Condition{Id: "7-21", Type: "training"}]).Should(Equal("1"))

	})

	BeforeEach(func() {
		fakePrometheus = util.NewFakePrometheusServer()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		fakePrometheus.Close()
		time.Sleep(100 * time.Millisecond)
	})
})
