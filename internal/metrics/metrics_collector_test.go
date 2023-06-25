package metrics

import (
	"context"
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

	})

	BeforeEach(func() {
		prom, _ := NewFakePrometheusServer()
		defer prom.Close()
		client, err := prometheusapi.NewClient(prometheusapi.Config{Address: prom.URL})
		Expect(err).NotTo(HaveOccurred())
		api := prometheusv1.NewAPI(client)
		collector, err := NewCollector(api)
		Expect(err).NotTo(HaveOccurred())

		go collector.Start(ctx)

	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
	})
})

func NewFakePrometheusServer() (*httptest.Server, error) {
	// var (
	// 	lastMethod string
	// 	lastBody   []byte
	// 	lastPath   string
	// )
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// lastMethod = r.Method
			var err error
			// lastBody, err = io.ReadAll(r.Body)
			if err != nil {

			}
			// lastPath = r.URL.EscapedPath()
			w.Header().Set("Content-Type", `text/plain; charset=utf-8`)
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			w.WriteHeader(http.StatusOK)

		}),
	), nil
}
