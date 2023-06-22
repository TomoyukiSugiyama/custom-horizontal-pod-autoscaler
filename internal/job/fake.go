package job

import (
	"context"

	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type FakeJobClient struct {
	ctrlClient ctrlClient.Client
	namespace  string
	name       string
}

var _ JobClient = (*FakeJobClient)(nil)

func FakeNew(ctrlClient ctrlClient.Client, namespace string, name string) *FakeJobClient {
	return &FakeJobClient{
		ctrlClient: ctrlClient,
		namespace:  namespace,
		name:       name,
	}
}

func (j *FakeJobClient) Start(ctx context.Context) {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	err := j.ctrlClient.Get(ctx, ctrlClient.ObjectKey{Namespace: j.namespace, Name: j.name}, &current)
	if err != nil {
		panic(1)
	}
	current.Status.DesiredMinReplicas = int32(1)
	current.Status.DesiredMaxReplicas = int32(5)
	j.ctrlClient.Status().Update(ctx, &current)
}

func (c *FakeJobClient) Stop() {
}
