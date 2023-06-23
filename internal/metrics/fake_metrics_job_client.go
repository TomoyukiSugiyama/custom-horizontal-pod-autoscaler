package metrics

import (
	"context"

	"k8s.io/utils/pointer"
	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

type FakeJobClient struct {
	desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec
}

// Guarantee *FakeJobClient implements JobClient.
var _ JobClient = (*FakeJobClient)(nil)

func FakeNew(desiredTemporaryScaleMetricSpec apiv1.TemporaryScaleMetricSpec) *FakeJobClient {
	return &FakeJobClient{
		desiredTemporaryScaleMetricSpec: desiredTemporaryScaleMetricSpec,
	}
}

func (j *FakeJobClient) Start(ctx context.Context) {
}

func (c *FakeJobClient) Stop() {
}

func (j *FakeJobClient) GetDesiredMinMaxReplicas() apiv1.TemporaryScaleMetricSpec {
	return apiv1.TemporaryScaleMetricSpec{
		MinReplicas: pointer.Int32(1),
		MaxReplicas: int32(5),
	}
}
