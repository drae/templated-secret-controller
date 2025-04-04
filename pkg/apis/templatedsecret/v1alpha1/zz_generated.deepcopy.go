//go:build !ignore_autogenerated
// +build !ignore_autogenerated

//
// Original source - secretgen-controller - Copyright 2024 The Carvel Authors.
// Re-organized and updated as - templated-secret-controller - (C) 2025 starstreak.dev
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	return
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
func (in *GenericStatus) DeepCopyInto(out *GenericStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericStatus.
func (in *GenericStatus) DeepCopy() *GenericStatus {
	if in == nil {
		return nil
	}
	out := new(GenericStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputResource) DeepCopyInto(out *InputResource) {
	*out = *in
	out.Ref = in.Ref
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputResource.
func (in *InputResource) DeepCopy() *InputResource {
	if in == nil {
		return nil
	}
	out := new(InputResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputResourceRef) DeepCopyInto(out *InputResourceRef) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputResourceRef.
func (in *InputResourceRef) DeepCopy() *InputResourceRef {
	if in == nil {
		return nil
	}
	out := new(InputResourceRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONPathTemplate) DeepCopyInto(out *JSONPathTemplate) {
	*out = *in
	if in.StringData != nil {
		in, out := &in.StringData, &out.StringData
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Metadata.DeepCopyInto(&out.Metadata)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONPathTemplate.
func (in *JSONPathTemplate) DeepCopy() *JSONPathTemplate {
	if in == nil {
		return nil
	}
	out := new(JSONPathTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplate) DeepCopyInto(out *SecretTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Metadata.DeepCopyInto(&out.Metadata)
	if in.StringData != nil {
		in, out := &in.StringData, &out.StringData
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplate.
func (in *SecretTemplate) DeepCopy() *SecretTemplate {
	if in == nil {
		return nil
	}
	out := new(SecretTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecretTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplateList) DeepCopyInto(out *SecretTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SecretTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplateList.
func (in *SecretTemplateList) DeepCopy() *SecretTemplateList {
	if in == nil {
		return nil
	}
	out := new(SecretTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecretTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplateMetadata) DeepCopyInto(out *SecretTemplateMetadata) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplateMetadata.
func (in *SecretTemplateMetadata) DeepCopy() *SecretTemplateMetadata {
	if in == nil {
		return nil
	}
	out := new(SecretTemplateMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplateSpec) DeepCopyInto(out *SecretTemplateSpec) {
	*out = *in
	if in.InputResources != nil {
		in, out := &in.InputResources, &out.InputResources
		*out = make([]InputResource, len(*in))
		copy(*out, *in)
	}
	if in.JSONPathTemplate != nil {
		in, out := &in.JSONPathTemplate, &out.JSONPathTemplate
		*out = new(JSONPathTemplate)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplateSpec.
func (in *SecretTemplateSpec) DeepCopy() *SecretTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(SecretTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretTemplateStatus) DeepCopyInto(out *SecretTemplateStatus) {
	*out = *in
	out.Secret = in.Secret
	in.GenericStatus.DeepCopyInto(&out.GenericStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretTemplateStatus.
func (in *SecretTemplateStatus) DeepCopy() *SecretTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(SecretTemplateStatus)
	in.DeepCopyInto(out)
	return out
}
