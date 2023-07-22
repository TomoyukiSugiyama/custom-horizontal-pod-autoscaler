package pusher

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ = Describe("Metrics Pusher", func() {
	var pusher MetricsPusher
	var err error
	It("Should get syncerTotal", func() {

		pusher.SetSyncerTotal(1)

		const metadata = `
		# HELP customhpa_syncer_total Total number of syncers per controller
		# TYPE customhpa_syncer_total gauge
		`
		expected := `

		customhpa_syncer_total{controller="customhorizontalpodautoscaler"} 1
		`
		err := testutil.CollectAndCompare(pusher.GetSyncerTotal(), strings.NewReader(metadata+expected), "customhpa_syncer_total")
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		pusher, err = NewPusher()
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
	})
})
