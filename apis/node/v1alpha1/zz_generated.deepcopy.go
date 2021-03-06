// +build !ignore_autogenerated

/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeJoinRequest) DeepCopyInto(out *NodeJoinRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeJoinRequest.
func (in *NodeJoinRequest) DeepCopy() *NodeJoinRequest {
	if in == nil {
		return nil
	}
	out := new(NodeJoinRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeJoinRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeJoinRequestList) DeepCopyInto(out *NodeJoinRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeJoinRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeJoinRequestList.
func (in *NodeJoinRequestList) DeepCopy() *NodeJoinRequestList {
	if in == nil {
		return nil
	}
	out := new(NodeJoinRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeJoinRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeJoinRequestSpec) DeepCopyInto(out *NodeJoinRequestSpec) {
	*out = *in
	if in.ExtraSANs != nil {
		in, out := &in.ExtraSANs, &out.ExtraSANs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeJoinRequestSpec.
func (in *NodeJoinRequestSpec) DeepCopy() *NodeJoinRequestSpec {
	if in == nil {
		return nil
	}
	out := new(NodeJoinRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeJoinRequestStatus) DeepCopyInto(out *NodeJoinRequestStatus) {
	*out = *in
	if in.VPN != nil {
		in, out := &in.VPN, &out.VPN
		*out = new(VPN)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(commonv1alpha1.ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeJoinRequestStatus.
func (in *NodeJoinRequestStatus) DeepCopy() *NodeJoinRequestStatus {
	if in == nil {
		return nil
	}
	out := new(NodeJoinRequestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPN) DeepCopyInto(out *VPN) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPN.
func (in *VPN) DeepCopy() *VPN {
	if in == nil {
		return nil
	}
	out := new(VPN)
	in.DeepCopyInto(out)
	return out
}
