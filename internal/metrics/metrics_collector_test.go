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
)

var _ = Describe("MetricsJobClient", func() {
	ctx := context.Background()

	It("Should get persedQueryResult", func() {
		fakePrometheus, _ := NewFakePrometheusServer()
		defer fakePrometheus.Close()

		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: fakePrometheus.URL})

		Expect(err).NotTo(HaveOccurred())
		api := prometheusv1.NewAPI(client)
		collector, err := NewCollector(api, WithMetricsCollectorInterval(50*time.Millisecond))
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		res := collector.GetPersedQueryResult()
		Expect(res[metricType{duration: "7-21", jobType: "training"}]).Should(Equal("1"))

	})

	BeforeEach(func() {
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
	})
})

func NewFakePrometheusServer() (*httptest.Server, error) {

	type apiResponse struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	data := []byte(
		`{
			"resultType":"vector",
			"result":[
				{
					"metric":{
						"__name__":"temporary_scale",
						"job":"prometheus",
						"instance":"localhost:9090",
						"exported_job":"temporary_scale_job_7-21_training",
						"duration":"7-21",
						"type":"training"
					},
					"value":[1435781451.781,"1"]
				}
			]
		}`,
	)

	resp := apiResponse{
		Status: "success",
		Data:   data,
	}

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			b, err := json.Marshal(resp)
			Expect(err).NotTo(HaveOccurred())
			w.Write(b)
			w.WriteHeader(http.StatusOK)
		}),
	), nil
}
