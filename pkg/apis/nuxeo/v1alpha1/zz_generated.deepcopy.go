// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DummyRevProxySpec) DeepCopyInto(out *DummyRevProxySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DummyRevProxySpec.
func (in *DummyRevProxySpec) DeepCopy() *DummyRevProxySpec {
	if in == nil {
		return nil
	}
	out := new(DummyRevProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NginxRevProxySpec) DeepCopyInto(out *NginxRevProxySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NginxRevProxySpec.
func (in *NginxRevProxySpec) DeepCopy() *NginxRevProxySpec {
	if in == nil {
		return nil
	}
	out := new(NginxRevProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeSet) DeepCopyInto(out *NodeSet) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = make([]NuxeoStorageSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.NuxeoConfig.DeepCopyInto(&out.NuxeoConfig)
	in.PodTemplate.DeepCopyInto(&out.PodTemplate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeSet.
func (in *NodeSet) DeepCopy() *NodeSet {
	if in == nil {
		return nil
	}
	out := new(NodeSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Nuxeo) DeepCopyInto(out *Nuxeo) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nuxeo.
func (in *Nuxeo) DeepCopy() *Nuxeo {
	if in == nil {
		return nil
	}
	out := new(Nuxeo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Nuxeo) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoAccess) DeepCopyInto(out *NuxeoAccess) {
	*out = *in
	out.TargetPort = in.TargetPort
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoAccess.
func (in *NuxeoAccess) DeepCopy() *NuxeoAccess {
	if in == nil {
		return nil
	}
	out := new(NuxeoAccess)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoConfig) DeepCopyInto(out *NuxeoConfig) {
	*out = *in
	if in.NuxeoTemplates != nil {
		in, out := &in.NuxeoTemplates, &out.NuxeoTemplates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NuxeoPackages != nil {
		in, out := &in.NuxeoPackages, &out.NuxeoPackages
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.NuxeoConf.DeepCopyInto(&out.NuxeoConf)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoConfig.
func (in *NuxeoConfig) DeepCopy() *NuxeoConfig {
	if in == nil {
		return nil
	}
	out := new(NuxeoConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoConfigSetting) DeepCopyInto(out *NuxeoConfigSetting) {
	*out = *in
	in.ValueFrom.DeepCopyInto(&out.ValueFrom)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoConfigSetting.
func (in *NuxeoConfigSetting) DeepCopy() *NuxeoConfigSetting {
	if in == nil {
		return nil
	}
	out := new(NuxeoConfigSetting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoList) DeepCopyInto(out *NuxeoList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Nuxeo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoList.
func (in *NuxeoList) DeepCopy() *NuxeoList {
	if in == nil {
		return nil
	}
	out := new(NuxeoList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NuxeoList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoSpec) DeepCopyInto(out *NuxeoSpec) {
	*out = *in
	out.RevProxy = in.RevProxy
	out.Service = in.Service
	out.Access = in.Access
	if in.NodeSets != nil {
		in, out := &in.NodeSets, &out.NodeSets
		*out = make([]NodeSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoSpec.
func (in *NuxeoSpec) DeepCopy() *NuxeoSpec {
	if in == nil {
		return nil
	}
	out := new(NuxeoSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoStatus) DeepCopyInto(out *NuxeoStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoStatus.
func (in *NuxeoStatus) DeepCopy() *NuxeoStatus {
	if in == nil {
		return nil
	}
	out := new(NuxeoStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuxeoStorageSpec) DeepCopyInto(out *NuxeoStorageSpec) {
	*out = *in
	in.VolumeClaimTemplate.DeepCopyInto(&out.VolumeClaimTemplate)
	in.VolumeSource.DeepCopyInto(&out.VolumeSource)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuxeoStorageSpec.
func (in *NuxeoStorageSpec) DeepCopy() *NuxeoStorageSpec {
	if in == nil {
		return nil
	}
	out := new(NuxeoStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RevProxySpec) DeepCopyInto(out *RevProxySpec) {
	*out = *in
	out.Nginx = in.Nginx
	out.Dummy = in.Dummy
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RevProxySpec.
func (in *RevProxySpec) DeepCopy() *RevProxySpec {
	if in == nil {
		return nil
	}
	out := new(RevProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}
