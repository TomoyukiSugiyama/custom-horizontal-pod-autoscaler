package metrics

import (
	"context"

	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

type FakeMetricsJobClient struct {
	desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec
}

// Guarantee *FakeMetricsJobClient implements JobClient.
var _ MetricsJobClient = (*FakeMetricsJobClient)(nil)

func FakeNew(desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec) *FakeMetricsJobClient {
	return &FakeMetricsJobClient{
		desiredTemporaryScaleMetricSpec: desiredTemporaryScaleMetricSpec,
	}
}

func (j *FakeMetricsJobClient) Start(ctx context.Context) {
}

func (j *FakeMetricsJobClient) Stop() {
}

func (j *FakeMetricsJobClient) GetDesiredMinMaxReplicas() apiv1.TemporaryScaleMetricSpec {
	return j.desiredTemporaryScaleMetricSpec
}
