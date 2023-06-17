/*
Copyright 2023.

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

package v1

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomHorizontalPodAutoscalerSpec defines the desired state of CustomHorizontalPodAutoscaler
type CustomHorizontalPodAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	HorizontalPodAutoscalerName string                                         `json:"horizontalPodAutoscalerName"`
	MinReplicas                 *int32                                         `json:"minReplicas"`
	MinReplicasTraining         *int32                                         `json:"minReplicasTraining"`
	MaxReplicas                 int32                                          `json:"maxReplicas"`
	MaxReplicasTraining         int32                                          `json:"maxReplicasTraining"`
	ScaleTargetRef              autoscalingv2.CrossVersionObjectReference      `json:"scaleTargetRef"`
	Metrics                     []autoscalingv2.MetricSpec                     `json:"metrics"`
	Behavior                    *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior"`
}

// CustomHorizontalPodAutoscalerStatus defines the observed state of CustomHorizontalPodAutoscaler
type CustomHorizontalPodAutoscalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ObservedGeneration *int64                                           `json:"observedGeneration"`
	LastScaleTime      *metav1.Time                                     `json:"lastScaleTime"`
	CurrentReplicas    int32                                            `json:"currentReplicas"`
	DesiredReplicas    int32                                            `json:"desiredReplicas"`
	CurrentMetrics     []autoscalingv2.MetricStatus                     `json:"currentMetrics"`
	Conditions         []autoscalingv2.HorizontalPodAutoscalerCondition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomHorizontalPodAutoscaler is the Schema for the customhorizontalpodautoscalers API
type CustomHorizontalPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomHorizontalPodAutoscalerSpec   `json:"spec,omitempty"`
	Status CustomHorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomHorizontalPodAutoscalerList contains a list of CustomHorizontalPodAutoscaler
type CustomHorizontalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomHorizontalPodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomHorizontalPodAutoscaler{}, &CustomHorizontalPodAutoscalerList{})
}
