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

type FakeSyncer struct {
	desiredConditionalReplicasSpec apiv1.ConditionalReplicasSpec
}

// Guarantee *FakeSyncer implements Syncer.
var _ Syncer = (*FakeSyncer)(nil)

func FakeNew(desiredConditionalReplicasSpec apiv1.ConditionalReplicasSpec) *FakeSyncer {
	return &FakeSyncer{
		desiredConditionalReplicasSpec: desiredConditionalReplicasSpec,
	}
}

func (s *FakeSyncer) Start(ctx context.Context) {
}

func (s *FakeSyncer) Stop() {
}

func (s *FakeSyncer) GetDesiredMinMaxReplicas() apiv1.ConditionalReplicasSpec {
	return s.desiredConditionalReplicasSpec
}
