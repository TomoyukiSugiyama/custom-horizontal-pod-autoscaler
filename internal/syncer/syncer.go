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
	s := &syncer{
		interval:                       30 * time.Second,
		stopCh:                         make(chan struct{}),
		metricsCollector:               metricsCollector,
		ctrlClient:                     ctrlClient,
		namespacedName:                 namespacedName,
		desiredConditionalReplicasSpec: customautoscalingv1.ConditionalReplicasSpec{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func WithSyncersInterval(interval time.Duration) Option {
	return func(s *syncer) {
		s.interval = interval
	}
}

func (s *syncer) getConditionalReplicasTarget(ctx context.Context) {
	logger := log.FromContext(ctx)
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	s.ctrlClient.Get(ctx, s.namespacedName, &current)
	s.updateDesiredMinMaxReplicas(ctx)

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
		"minReplicas", s.desiredConditionalReplicasSpec.MinReplicas,
		"maxReplicas", s.desiredConditionalReplicasSpec.MaxReplicas,
	)
}

func (s *syncer) updateDesiredMinMaxReplicas(ctx context.Context) {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	s.ctrlClient.Get(ctx, s.namespacedName, &current)

	s.desiredConditionalReplicasSpec.MinReplicas = pointer.Int32(1)
	if current.Spec.MinReplicas != nil {
		s.desiredConditionalReplicasSpec.MinReplicas = current.Spec.MinReplicas
	}
	s.desiredConditionalReplicasSpec.MaxReplicas = &current.Spec.MaxReplicas

	for _, target := range current.Spec.ConditionalReplicasSpecs {
		k := customautoscalingv1.Condition{
			Type: target.Condition.Type,
			Id:   target.Condition.Id,
		}
		res := s.metricsCollector.GetPersedQueryResult()
		v, isExist := res[k]
		if isExist && v == "1" && *s.desiredConditionalReplicasSpec.MaxReplicas <= *target.MaxReplicas {
			s.desiredConditionalReplicasSpec.MinReplicas = pointer.Int32(1)
			if target.MinReplicas != nil {
				s.desiredConditionalReplicasSpec.MinReplicas = target.MinReplicas
			}
			s.desiredConditionalReplicasSpec.MaxReplicas = target.MaxReplicas
		}
	}
}

func (s *syncer) updateStatus(ctx context.Context) error {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	s.ctrlClient.Get(ctx, s.namespacedName, &current)
	current.Status.DesiredMinReplicas = *s.desiredConditionalReplicasSpec.MinReplicas
	current.Status.DesiredMaxReplicas = *s.desiredConditionalReplicasSpec.MaxReplicas
	return s.ctrlClient.Status().Update(ctx, &current)
}

func (s *syncer) GetDesiredMinMaxReplicas() customautoscalingv1.ConditionalReplicasSpec {
	return s.desiredConditionalReplicasSpec
}

func (s *syncer) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting syncer")
	defer logger.Info("shut down syncer")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("received scheduler tick")
			ctx = context.Background()
			s.getConditionalReplicasTarget(ctx)
			if err := s.updateStatus(ctx); err != nil {
				logger.Error(err, "unable to update status")
			}
		case <-s.stopCh:
			logger.Info("received stop signal")
			return
		}
	}
}

func (s *syncer) Stop() {
	close(s.stopCh)
}
