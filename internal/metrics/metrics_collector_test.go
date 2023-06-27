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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prometheusapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

var _ = Describe("Syncer", func() {
	ctx := context.Background()

	It("Should get persedQueryResult", func() {
		fakePrometheus := NewFakePrometheusServer()
		defer fakePrometheus.Close()

		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: fakePrometheus.URL})

		Expect(err).NotTo(HaveOccurred())
		api := prometheusv1.NewAPI(client)
		collector, err := NewCollector(api, WithMetricsCollectorInterval(50*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		res := collector.GetPersedQueryResult()
		Expect(res[customautoscalingv1.Condition{Id: "7-21", Type: "training"}]).Should(Equal("1"))

	})

	BeforeEach(func() {
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
	})
})

func NewFakePrometheusServer() *httptest.Server {

	type metric struct {
		Name        string `json:"__name__"`
		Job         string `json:"job"`
		Instance    string `json:"instance"`
		ExportedJob string `json:"exported_job"`
		Id          string `json:"id"`
		Type        string `json:"type"`
	}

	type result struct {
		Metric metric          `json:"metric"`
		Value  json.RawMessage `json:"value"`
	}

	type data struct {
		ResultType string   `json:"resultType"`
		Result     []result `json:"result"`
	}

	type apiResponse struct {
		Status string `json:"status"`
		Data   data   `json:"data"`
	}

	res := apiResponse{
		Status: "success",
		Data: data{
			ResultType: "vector",
			Result: []result{
				{
					Metric: metric{
						Name:        "temporary_scale",
						Job:         "prometheus",
						Instance:    "localhost:9090",
						ExportedJob: "temporary_scale_job_7-21_training",
						Id:          "7-21",
						Type:        "training",
					},
					Value: []byte(`[1435781451.781,"1"]`),
				},
			},
		},
	}

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			b, err := json.Marshal(res)
			Expect(err).NotTo(HaveOccurred())
			w.Write(b)
			w.WriteHeader(http.StatusOK)
		}),
	)
}
