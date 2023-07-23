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

package pusher

import (
	"strings"
	"time"

	customautoscalingv1alpha1 "github.com/TomoyukiSugiyama/custom-horizontal-pod-autoscaler/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ = Describe("Metrics Pusher", func() {
	pusher, err := NewPusher()
	It("Should get syncerTotal", func() {
		pusher.SetSyncerTotal(1)

		const metadata = `
		# HELP customhpa_syncer_total Total number of syncers per controller
		# TYPE customhpa_syncer_total gauge
		`
		expected := `

		customhpa_syncer_total{controller="customhorizontalpodautoscaler"} 1
		`
		err = testutil.CollectAndCompare(pusher.GetSyncerTotal(), strings.NewReader(metadata+expected), "customhpa_syncer_total")
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should get collectorNotReady", func() {
		pusher.SetCollectorStatus("test-ns", "test-custom-hpa", customautoscalingv1alpha1.CollectorNotReady)

		const metadataNotReady = `
		# HELP customhpa_collector_notready The controller status about not ready condition
		# TYPE customhpa_collector_notready gauge
		`
		expectedNotReady := `

		customhpa_collector_notready{controller="customhorizontalpodautoscaler",name="test-custom-hpa",namespace="test-ns"} 1
		`
		err = testutil.CollectAndCompare(pusher.GetCollectorNotReady(), strings.NewReader(metadataNotReady+expectedNotReady), "customhpa_collector_notready")
		Expect(err).NotTo(HaveOccurred())

		const metadataAvailable = `
		# HELP customhpa_collector_available The controller status about available condition
		# TYPE customhpa_collector_available gauge
		`
		expectedAvailable := `

		customhpa_collector_available{controller="customhorizontalpodautoscaler",name="test-custom-hpa",namespace="test-ns"} 0
		`
		err = testutil.CollectAndCompare(pusher.GetCollectorAvailable(), strings.NewReader(metadataAvailable+expectedAvailable), "customhpa_collector_notready")
		Expect(err).NotTo(HaveOccurred())

	})

	It("Should get collectorAvailable", func() {
		pusher.SetCollectorStatus("test-ns", "test-custom-hpa", customautoscalingv1alpha1.CollectorAvailable)

		const metadataNotReady = `
		# HELP customhpa_collector_notready The controller status about not ready condition
		# TYPE customhpa_collector_notready gauge
		`
		expectedNotReady := `

		customhpa_collector_notready{controller="customhorizontalpodautoscaler",name="test-custom-hpa",namespace="test-ns"} 0
		`
		err = testutil.CollectAndCompare(pusher.GetCollectorNotReady(), strings.NewReader(metadataNotReady+expectedNotReady), "customhpa_collector_notready")
		Expect(err).NotTo(HaveOccurred())

		const metadataAvailable = `
		# HELP customhpa_collector_available The controller status about available condition
		# TYPE customhpa_collector_available gauge
		`
		expectedAvailable := `

		customhpa_collector_available{controller="customhorizontalpodautoscaler",name="test-custom-hpa",namespace="test-ns"} 1
		`
		err = testutil.CollectAndCompare(pusher.GetCollectorAvailable(), strings.NewReader(metadataAvailable+expectedAvailable), "customhpa_collector_notready")
		Expect(err).NotTo(HaveOccurred())

	})

	BeforeEach(func() {
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {

		time.Sleep(100 * time.Millisecond)
	})
})
