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

package metrics

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	apiv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	customautoscalingv1 "sample.com/custom-horizontal-pod-autoscaler/api/v1"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MetricsJobClient interface {
	Start(ctx context.Context)
	Stop()
	GetDesiredMinMaxReplicas() apiv1.ConditionalReplicasTargetSpec
}

type metricsJobClient struct {
	metricsCollector                 MetricsCollector
	ctrlClient                       ctrlClient.Client
	interval                         time.Duration
	stopCh                           chan struct{}
	desiredConditionalReplicasTarget apiv1.ConditionalReplicasTargetSpec
	namespacedName                   types.NamespacedName
}

var _ MetricsJobClient = (*metricsJobClient)(nil)

type Option func(*metricsJobClient)

func New(metricsCollector MetricsCollector, ctrlClient ctrlClient.Client, namespacedName types.NamespacedName, opts ...Option) (MetricsJobClient, error) {
	j := &metricsJobClient{
		interval:                         30 * time.Second,
		stopCh:                           make(chan struct{}),
		metricsCollector:                 metricsCollector,
		ctrlClient:                       ctrlClient,
		namespacedName:                   namespacedName,
		desiredConditionalReplicasTarget: apiv1.ConditionalReplicasTargetSpec{},
	}

	for _, opt := range opts {
		opt(j)
	}

	return j, nil
}

func WithMetricsJobClientsInterval(interval time.Duration) Option {
	return func(j *metricsJobClient) {
		j.interval = interval
	}
}

func (j *metricsJobClient) getConditionalReplicasTarget(ctx context.Context) {
	logger := log.FromContext(ctx)
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	j.updateDesiredMinMaxReplicas(ctx)

	for _, target := range current.Spec.ConditionalReplicasTargets {
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
		"minReplicas", j.desiredConditionalReplicasTarget.MinReplicas,
		"maxReplicas", j.desiredConditionalReplicasTarget.MaxReplicas,
	)
}

func (j *metricsJobClient) updateDesiredMinMaxReplicas(ctx context.Context) {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)

	j.desiredConditionalReplicasTarget.MinReplicas = pointer.Int32(1)
	if current.Spec.MinReplicas != nil {
		j.desiredConditionalReplicasTarget.MinReplicas = current.Spec.MinReplicas
	}
	j.desiredConditionalReplicasTarget.MaxReplicas = &current.Spec.MaxReplicas

	for _, target := range current.Spec.ConditionalReplicasTargets {
		k := metricType{
			jobType:  target.Condition.Type,
			duration: target.Condition.Id,
		}
		res := j.metricsCollector.GetPersedQueryResult()
		v, isExist := res[k]
		if isExist && v == "1" && *j.desiredConditionalReplicasTarget.MaxReplicas <= *target.MaxReplicas {
			j.desiredConditionalReplicasTarget.MinReplicas = pointer.Int32(1)
			if target.MinReplicas != nil {
				j.desiredConditionalReplicasTarget.MinReplicas = target.MinReplicas
			}
			j.desiredConditionalReplicasTarget.MaxReplicas = target.MaxReplicas
		}
	}
}

func (j *metricsJobClient) updateStatus(ctx context.Context) error {
	var current customautoscalingv1.CustomHorizontalPodAutoscaler
	j.ctrlClient.Get(ctx, j.namespacedName, &current)
	current.Status.DesiredMinReplicas = *j.desiredConditionalReplicasTarget.MinReplicas
	current.Status.DesiredMaxReplicas = *j.desiredConditionalReplicasTarget.MaxReplicas
	return j.ctrlClient.Status().Update(ctx, &current)
}

func (j *metricsJobClient) GetDesiredMinMaxReplicas() apiv1.ConditionalReplicasTargetSpec {
	return j.desiredConditionalReplicasTarget
}

func (j *metricsJobClient) Start(ctx context.Context) {
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

func (j *metricsJobClient) Stop() {
	close(j.stopCh)
}
