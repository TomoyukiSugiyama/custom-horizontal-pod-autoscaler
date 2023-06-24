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

type TemporaryScaleMetricSpec struct {
	// Type is the name of the target metric that determines the upper and lower limits
	// of the number of pods during a temporary scale.
	Type string `json:"type"`
	// duration is a label indicating the duration of the target metric that determines
	// the upper and lower limits of the number of pods during the temporary scale.
	Duration string `json:"duration"`
	// minReplicas is the lower limit for the number of replicas to which the custom autoscaler
	// can scale up according to the temporary scale metrics. It defaults to 1 pod.
	// +optional
	MinReplicas *int32 `json:"minReplicas"`
	// maxReplicas is the upper limit for the number of replicas to which the custom autoscaler
	// can scale up according to the temporary scale metrics. It cannot be less that minReplicas.
	MaxReplicas int32 `json:"maxReplicas"`
}

// CustomHorizontalPodAutoscalerSpec defines the desired state of CustomHorizontalPodAutoscaler
type CustomHorizontalPodAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// scaleTargetRef points to the target resource to scale, and is used to the pods for which metrics
	// should be collected, as well as to actually change the replica count.
	ScaleTargetRef              autoscalingv2.CrossVersionObjectReference `json:"scaleTargetRef" protobuf:"bytes,1,opt,name=scaleTargetRef"`
	HorizontalPodAutoscalerName string                                    `json:"horizontalPodAutoscalerName"`
	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// If not set, the default metric will be set to 80% average CPU utilization.
	// +listType=atomic
	// +optional
	Metrics []autoscalingv2.MetricSpec `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
	// If temporaryScaleMetrics is set, the minReplicas and maxReplicas of temporaryScaleMetrics are used
	// in preference to the minReplicas and maxReplicas of spec only when the corresponding metrics
	// for name and duration of temporaryScaleMetrics are 1.
	// If not set, the default minReplicas and maxReplicas of spec are used.
	// +listType=atomic
	// +optional
	TemporaryScaleMetrics []TemporaryScaleMetricSpec `json:"temporaryScaleMetrics"`
	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty" protobuf:"bytes,5,opt,name=behavior"`
}

// CustomHorizontalPodAutoscalerStatus defines the observed state of CustomHorizontalPodAutoscaler
type CustomHorizontalPodAutoscalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// observedGeneration is the most recent generation observed by this autoscaler.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// lastScaleTime is the last time the HorizontalPodAutoscaler scaled the number of pods,
	// used by the autoscaler to control how often the number of pods is changed.
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty" protobuf:"bytes,2,opt,name=lastScaleTime"`
	// currentReplicas is current number of replicas of pods managed by this autoscaler,
	// as last seen by the autoscaler.
	// +optional
	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,3,opt,name=currentReplicas"`
	// desiredReplicas is the desired number of replicas of pods managed by this autoscaler,
	// as last calculated by the autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas" protobuf:"varint,4,opt,name=desiredReplicas"`
	// currentMetrics is the last read state of the metrics used by this autoscaler.
	// +listType=atomic
	// +optional
	CurrentMetrics []autoscalingv2.MetricStatus `json:"currentMetrics" protobuf:"bytes,5,rep,name=currentMetrics"`
	// conditions is the set of conditions required for this autoscaler to scale its target,
	// and indicates whether or not those conditions are met.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []autoscalingv2.HorizontalPodAutoscalerCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" listType:"map" protobuf:"bytes,6,rep,name=conditions"`
	// currentMinReplicas is current lower limit for the number of replicas to which the autoscaler can scale down,
	// as last seen by the custom autoscaler.
	// +optional
	CurrentMinReplicas int32 `json:"currentMinReplicas"`
	// currentMaxReplicas is current upper limit for the number of replicas to which the autoscaler can scale up,
	// as last seen by the custom autoscaler.
	// +optional
	CueerntMaxReplicas int32 `json:"currentMaxReplicas"`
	// desiredMinReplicas is desired lower limit for the number of replicas to which the autoscaler can scale down,
	// as last calculated by the custom autoscaler.
	// +optional
	DesiredMinReplicas int32 `json:"desiredMinReplicas"`
	// desiredMaxReplicas is desired upper limit for the number of replicas to which the autoscaler can scale up,
	// as last calculated by the custom autoscaler.
	// +optional
	DesiredMaxReplicas int32 `json:"desiredMaxReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Reference",type="string",JSONPath=".spec.horizontalPodAutoscalerName"
//+kubebuilder:printcolumn:name="Minpod",type="integer",JSONPath=".status.currentMinReplicas"
//+kubebuilder:printcolumn:name="Maxpod",type="integer",JSONPath=".status.currentMaxReplicas"
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.currentReplicas"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
