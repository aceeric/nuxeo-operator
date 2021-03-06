// +build !ignore_autogenerated

/*
Copyright 2020 Eric Ace.

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
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackingService) DeepCopyInto(out *BackingService) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]BackingServiceResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Preconfigured.DeepCopyInto(&out.Preconfigured)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackingService.
func (in *BackingService) DeepCopy() *BackingService {
	if in == nil {
		return nil
	}
	out := new(BackingService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackingServiceResource) DeepCopyInto(out *BackingServiceResource) {
	*out = *in
	out.GroupVersionKind = in.GroupVersionKind
	if in.Projections != nil {
		in, out := &in.Projections, &out.Projections
		*out = make([]ResourceProjection, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackingServiceResource.
func (in *BackingServiceResource) DeepCopy() *BackingServiceResource {
	if in == nil {
		return nil
	}
	out := new(BackingServiceResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertTransform) DeepCopyInto(out *CertTransform) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertTransform.
func (in *CertTransform) DeepCopy() *CertTransform {
	if in == nil {
		return nil
	}
	out := new(CertTransform)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Contribution) DeepCopyInto(out *Contribution) {
	*out = *in
	if in.Templates != nil {
		in, out := &in.Templates, &out.Templates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.VolumeSource.DeepCopyInto(&out.VolumeSource)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Contribution.
func (in *Contribution) DeepCopy() *Contribution {
	if in == nil {
		return nil
	}
	out := new(Contribution)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NginxRevProxySpec) DeepCopyInto(out *NginxRevProxySpec) {
	*out = *in
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
	in.Resources.DeepCopyInto(&out.Resources)
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
	if in.Contributions != nil {
		in, out := &in.Contributions, &out.Contributions
		*out = make([]Contribution, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
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
	if in.OfflinePackages != nil {
		in, out := &in.OfflinePackages, &out.OfflinePackages
		*out = make([]OfflinePackage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
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
	if in.BackingServices != nil {
		in, out := &in.BackingServices, &out.BackingServices
		*out = make([]BackingService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]v1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
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
func (in *OfflinePackage) DeepCopyInto(out *OfflinePackage) {
	*out = *in
	in.ValueFrom.DeepCopyInto(&out.ValueFrom)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OfflinePackage.
func (in *OfflinePackage) DeepCopy() *OfflinePackage {
	if in == nil {
		return nil
	}
	out := new(OfflinePackage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PreconfiguredBackingService) DeepCopyInto(out *PreconfiguredBackingService) {
	*out = *in
	if in.Settings != nil {
		in, out := &in.Settings, &out.Settings
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PreconfiguredBackingService.
func (in *PreconfiguredBackingService) DeepCopy() *PreconfiguredBackingService {
	if in == nil {
		return nil
	}
	out := new(PreconfiguredBackingService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceProjection) DeepCopyInto(out *ResourceProjection) {
	*out = *in
	out.Transform = in.Transform
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceProjection.
func (in *ResourceProjection) DeepCopy() *ResourceProjection {
	if in == nil {
		return nil
	}
	out := new(ResourceProjection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RevProxySpec) DeepCopyInto(out *RevProxySpec) {
	*out = *in
	out.Nginx = in.Nginx
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
