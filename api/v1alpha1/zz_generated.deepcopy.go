//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/autoscaling/v2"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConditionalReplicasSpec) DeepCopyInto(out *ConditionalReplicasSpec) {
	*out = *in
	out.Condition = in.Condition
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.MaxReplicas != nil {
		in, out := &in.MaxReplicas, &out.MaxReplicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConditionalReplicasSpec.
func (in *ConditionalReplicasSpec) DeepCopy() *ConditionalReplicasSpec {
	if in == nil {
		return nil
	}
	out := new(ConditionalReplicasSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomHorizontalPodAutoscaler) DeepCopyInto(out *CustomHorizontalPodAutoscaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomHorizontalPodAutoscaler.
func (in *CustomHorizontalPodAutoscaler) DeepCopy() *CustomHorizontalPodAutoscaler {
	if in == nil {
		return nil
	}
	out := new(CustomHorizontalPodAutoscaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomHorizontalPodAutoscaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomHorizontalPodAutoscalerList) DeepCopyInto(out *CustomHorizontalPodAutoscalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomHorizontalPodAutoscaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomHorizontalPodAutoscalerList.
func (in *CustomHorizontalPodAutoscalerList) DeepCopy() *CustomHorizontalPodAutoscalerList {
	if in == nil {
		return nil
	}
	out := new(CustomHorizontalPodAutoscalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomHorizontalPodAutoscalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomHorizontalPodAutoscalerSpec) DeepCopyInto(out *CustomHorizontalPodAutoscalerSpec) {
	*out = *in
	out.ScaleTargetRef = in.ScaleTargetRef
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]v2.MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConditionalReplicasSpecs != nil {
		in, out := &in.ConditionalReplicasSpecs, &out.ConditionalReplicasSpecs
		*out = make([]ConditionalReplicasSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Behavior != nil {
		in, out := &in.Behavior, &out.Behavior
		*out = new(v2.HorizontalPodAutoscalerBehavior)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomHorizontalPodAutoscalerSpec.
func (in *CustomHorizontalPodAutoscalerSpec) DeepCopy() *CustomHorizontalPodAutoscalerSpec {
	if in == nil {
		return nil
	}
	out := new(CustomHorizontalPodAutoscalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomHorizontalPodAutoscalerStatus) DeepCopyInto(out *CustomHorizontalPodAutoscalerStatus) {
	*out = *in
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = **in
	}
	if in.LastScaleTime != nil {
		in, out := &in.LastScaleTime, &out.LastScaleTime
		*out = (*in).DeepCopy()
	}
	if in.CurrentMetrics != nil {
		in, out := &in.CurrentMetrics, &out.CurrentMetrics
		*out = make([]v2.MetricStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v2.HorizontalPodAutoscalerCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomHorizontalPodAutoscalerStatus.
func (in *CustomHorizontalPodAutoscalerStatus) DeepCopy() *CustomHorizontalPodAutoscalerStatus {
	if in == nil {
		return nil
	}
	out := new(CustomHorizontalPodAutoscalerStatus)
	in.DeepCopyInto(out)
	return out
}
