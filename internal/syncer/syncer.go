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
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	metricspkg "sample.com/custom-horizontal-pod-autoscaler/internal/metrics"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Syncer interface {
	Start(ctx context.Context)
	Stop()
	GetDesiredMinMaxReplicas() customautoscalingv1.ConditionalReplicasSpec
}

type syncer struct {
	metricsCollector               metricspkg.MetricsCollector
	ctrlClient                     ctrlClient.Client
	interval                       time.Duration
	stopCh                         chan struct{}
	desiredConditionalReplicasSpec customautoscalingv1.ConditionalReplicasSpec
	namespacedName                 types.NamespacedName
}

var _ Syncer = (*syncer)(nil)

type Option func(*syncer)

func New(metricsCollector metricspkg.MetricsCollector, ctrlClient ctrlClient.Client, namespacedName types.NamespacedName, opts ...Option) (Syncer, error) {
	j := &syncer{
		interval:                       30 * time.Second,
		stopCh:                         make(chan struct{}),
		metricsCollector:               metricsCollector,
		ctrlClient:                     ctrlClient,
		namespacedName:                 namespacedName,
		desiredConditionalReplicasSpec: customautoscalingv1.ConditionalReplicasSpec{},
	}

	for _, opt := range opts {
		opt(j)
	}

	return j, nil
}

func WithSyncersInterval(interval time.Duration) Option {
	return func(j *syncer) {
		j.interval = interval
	}
}

func (j *syncer) getConditionalReplicasTarget(ctx context.Context) {
	logger := log.FromContext(ctx)
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	j.updateDesiredMinMaxReplicas(ctx)

	for _, target := range current.Spec.ConditionalReplicasSpecs {
		logger.Info(
			"customHPA settings",
			"id", target.Condition.Id,
			"type", target.Condition.Type,
			"minReplicas", target.MinReplicas,
			"maxReplicas", target.MaxReplicas,
		)
	}

	logger.Info(
		"update desired min and max replicas",
		"minReplicas", j.desiredConditionalReplicasSpec.MinReplicas,
		"maxReplicas", j.desiredConditionalReplicasSpec.MaxReplicas,
	)
}

func (j *syncer) updateDesiredMinMaxReplicas(ctx context.Context) {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)

	j.desiredConditionalReplicasSpec.MinReplicas = pointer.Int32(1)
	if current.Spec.MinReplicas != nil {
		j.desiredConditionalReplicasSpec.MinReplicas = current.Spec.MinReplicas
	}
	j.desiredConditionalReplicasSpec.MaxReplicas = &current.Spec.MaxReplicas

	for _, target := range current.Spec.ConditionalReplicasSpecs {
		k := metricspkg.MetricType{
			JobType: target.Condition.Type,
			Id:      target.Condition.Id,
		}
		res := j.metricsCollector.GetPersedQueryResult()
		v, isExist := res[k]
		if isExist && v == "1" && *j.desiredConditionalReplicasSpec.MaxReplicas <= *target.MaxReplicas {
			j.desiredConditionalReplicasSpec.MinReplicas = pointer.Int32(1)
			if target.MinReplicas != nil {
				j.desiredConditionalReplicasSpec.MinReplicas = target.MinReplicas
			}
			j.desiredConditionalReplicasSpec.MaxReplicas = target.MaxReplicas
		}
	}
}

func (j *syncer) updateStatus(ctx context.Context) error {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	current.Status.DesiredMinReplicas = *j.desiredConditionalReplicasSpec.MinReplicas
	current.Status.DesiredMaxReplicas = *j.desiredConditionalReplicasSpec.MaxReplicas
	return j.ctrlClient.Status().Update(ctx, &current)
}

func (j *syncer) GetDesiredMinMaxReplicas() customautoscalingv1.ConditionalReplicasSpec {
	return j.desiredConditionalReplicasSpec
}

func (j *syncer) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting job")
	defer logger.Info("shut down job")

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("received scheduler tick")
			ctx = context.Background()
			j.getConditionalReplicasTarget(ctx)
			if err := j.updateStatus(ctx); err != nil {
				logger.Error(err, "unable to update status")
			}
		case <-j.stopCh:
			logger.Info("received stop signal")
			return
		}
	}
}

func (j *syncer) Stop() {
	close(j.stopCh)
}
