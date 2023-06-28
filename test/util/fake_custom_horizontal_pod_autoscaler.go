package util

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	customautoscalingv1alpha1 "sample.com/custom-horizontal-pod-autoscaler/api/v1alpha1"
)

func NewCustomHorizontalPodAutoscaler() *customautoscalingv1alpha1.CustomHorizontalPodAutoscaler {
	minReplicas := int32(1)
	maxReplicas := int32(5)
	scaleTargetRef := autoscalingv2.CrossVersionObjectReference{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       "test-deployment",
	}
	metrics := []autoscalingv2.MetricSpec{
		{
			Type: "Pods",
			Pods: &autoscalingv2.PodsMetricSource{
				Metric: autoscalingv2.MetricIdentifier{
					Name: "memory_usage_bytes",
				},
				Target: autoscalingv2.MetricTarget{
					Type:         "AverageValue",
					AverageValue: resource.NewQuantity(8*1024*1024, resource.BinarySI),
				},
			},
		},
	}
	workdayMinRelpicas := int32(2)
	workdayMaxRelpicas := int32(4)
	trainingMinRelpicas := int32(5)
	trainingMaxRelpicas := int32(10)
	conditionalReplicasSpecs := []customautoscalingv1alpha1.ConditionalReplicasSpec{
		{
			Condition: customautoscalingv1alpha1.Condition{
				Type: "workday",
				Id:   "7-21",
			},
			MinReplicas: &workdayMinRelpicas,
			MaxReplicas: &workdayMaxRelpicas,
		},
		{
			Condition: customautoscalingv1alpha1.Condition{
				Type: "training",
				Id:   "7-21",
			},
			MinReplicas: &trainingMinRelpicas,
			MaxReplicas: &trainingMaxRelpicas,
		},
	}

	return &customautoscalingv1alpha1.CustomHorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-customhpa",
			Namespace: "dummy-namespace",
		},
		Spec: customautoscalingv1alpha1.CustomHorizontalPodAutoscalerSpec{
			HorizontalPodAutoscalerName: "test-hpa",
			MinReplicas:                 &minReplicas,
			MaxReplicas:                 maxReplicas,
			ScaleTargetRef:              scaleTargetRef,
			Metrics:                     metrics,
			ConditionalReplicasSpecs:    conditionalReplicasSpecs,
		},
	}
}
