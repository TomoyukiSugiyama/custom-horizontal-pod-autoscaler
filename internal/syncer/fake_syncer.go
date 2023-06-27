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

	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
)

type FakeMetricsJobClient struct {
	desiredConditionalReplicasSpec apiv1.ConditionalReplicasSpec
}

// Guarantee *FakeMetricsJobClient implements JobClient.
var _ MetricsJobClient = (*FakeMetricsJobClient)(nil)

func FakeNew(desiredConditionalReplicasSpec apiv1.ConditionalReplicasSpec) *FakeMetricsJobClient {
	return &FakeMetricsJobClient{
		desiredConditionalReplicasSpec: desiredConditionalReplicasSpec,
	}
}

func (j *FakeMetricsJobClient) Start(ctx context.Context) {
}

func (j *FakeMetricsJobClient) Stop() {
}

func (j *FakeMetricsJobClient) GetDesiredMinMaxReplicas() apiv1.ConditionalReplicasSpec {
	return j.desiredConditionalReplicasSpec
}
