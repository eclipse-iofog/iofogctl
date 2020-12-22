// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package apps

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentConfiguration) DeepCopyInto(out *AgentConfiguration) {
	*out = *in
	if in.DockerURL != nil {
		in, out := &in.DockerURL, &out.DockerURL
		*out = new(string)
		**out = **in
	}
	if in.DiskLimit != nil {
		in, out := &in.DiskLimit, &out.DiskLimit
		*out = new(int64)
		**out = **in
	}
	if in.DiskDirectory != nil {
		in, out := &in.DiskDirectory, &out.DiskDirectory
		*out = new(string)
		**out = **in
	}
	if in.MemoryLimit != nil {
		in, out := &in.MemoryLimit, &out.MemoryLimit
		*out = new(int64)
		**out = **in
	}
	if in.CPULimit != nil {
		in, out := &in.CPULimit, &out.CPULimit
		*out = new(int64)
		**out = **in
	}
	if in.LogLimit != nil {
		in, out := &in.LogLimit, &out.LogLimit
		*out = new(int64)
		**out = **in
	}
	if in.LogDirectory != nil {
		in, out := &in.LogDirectory, &out.LogDirectory
		*out = new(string)
		**out = **in
	}
	if in.LogFileCount != nil {
		in, out := &in.LogFileCount, &out.LogFileCount
		*out = new(int64)
		**out = **in
	}
	if in.StatusFrequency != nil {
		in, out := &in.StatusFrequency, &out.StatusFrequency
		*out = new(float64)
		**out = **in
	}
	if in.ChangeFrequency != nil {
		in, out := &in.ChangeFrequency, &out.ChangeFrequency
		*out = new(float64)
		**out = **in
	}
	if in.DeviceScanFrequency != nil {
		in, out := &in.DeviceScanFrequency, &out.DeviceScanFrequency
		*out = new(float64)
		**out = **in
	}
	if in.BluetoothEnabled != nil {
		in, out := &in.BluetoothEnabled, &out.BluetoothEnabled
		*out = new(bool)
		**out = **in
	}
	if in.WatchdogEnabled != nil {
		in, out := &in.WatchdogEnabled, &out.WatchdogEnabled
		*out = new(bool)
		**out = **in
	}
	if in.AbstractedHardwareEnabled != nil {
		in, out := &in.AbstractedHardwareEnabled, &out.AbstractedHardwareEnabled
		*out = new(bool)
		**out = **in
	}
	if in.RouterMode != nil {
		in, out := &in.RouterMode, &out.RouterMode
		*out = new(string)
		**out = **in
	}
	if in.RouterPort != nil {
		in, out := &in.RouterPort, &out.RouterPort
		*out = new(int)
		**out = **in
	}
	if in.UpstreamRouters != nil {
		in, out := &in.UpstreamRouters, &out.UpstreamRouters
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.NetworkRouter != nil {
		in, out := &in.NetworkRouter, &out.NetworkRouter
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentConfiguration.
func (in *AgentConfiguration) DeepCopy() *AgentConfiguration {
	if in == nil {
		return nil
	}
	out := new(AgentConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	if in.Microservices != nil {
		in, out := &in.Microservices, &out.Microservices
		*out = make([]Microservice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		copy(*out, *in)
	}
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(ApplicationTemplate)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Application.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationTemplate) DeepCopyInto(out *ApplicationTemplate) {
	*out = *in
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make([]TemplateVariable, len(*in))
		copy(*out, *in)
	}
	if in.Application != nil {
		in, out := &in.Application, &out.Application
		*out = new(ApplicationTemplateInfo)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationTemplate.
func (in *ApplicationTemplate) DeepCopy() *ApplicationTemplate {
	if in == nil {
		return nil
	}
	out := new(ApplicationTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationTemplateInfo) DeepCopyInto(out *ApplicationTemplateInfo) {
	*out = *in
	if in.Microservices != nil {
		in, out := &in.Microservices, &out.Microservices
		*out = make([]Microservice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationTemplateInfo.
func (in *ApplicationTemplateInfo) DeepCopy() *ApplicationTemplateInfo {
	if in == nil {
		return nil
	}
	out := new(ApplicationTemplateInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Applications) DeepCopyInto(out *Applications) {
	*out = *in
	if in.Applications != nil {
		in, out := &in.Applications, &out.Applications
		*out = make([]Application, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Applications.
func (in *Applications) DeepCopy() *Applications {
	if in == nil {
		return nil
	}
	out := new(Applications)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CatalogItem) DeepCopyInto(out *CatalogItem) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CatalogItem.
func (in *CatalogItem) DeepCopy() *CatalogItem {
	if in == nil {
		return nil
	}
	out := new(CatalogItem)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeaderMetadata) DeepCopyInto(out *HeaderMetadata) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderMetadata.
func (in *HeaderMetadata) DeepCopy() *HeaderMetadata {
	if in == nil {
		return nil
	}
	out := new(HeaderMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IofogController) DeepCopyInto(out *IofogController) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IofogController.
func (in *IofogController) DeepCopy() *IofogController {
	if in == nil {
		return nil
	}
	out := new(IofogController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Microservice) DeepCopyInto(out *Microservice) {
	*out = *in
	out.Agent = in.Agent
	if in.Images != nil {
		in, out := &in.Images, &out.Images
		*out = new(MicroserviceImages)
		**out = **in
	}
	in.Container.DeepCopyInto(&out.Container)
	out.Config = in.Config.DeepCopy()
	if in.Flow != nil {
		in, out := &in.Flow, &out.Flow
		*out = new(string)
		**out = **in
	}
	if in.Application != nil {
		in, out := &in.Application, &out.Application
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Microservice.
func (in *Microservice) DeepCopy() *Microservice {
	if in == nil {
		return nil
	}
	out := new(Microservice)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceAgent) DeepCopyInto(out *MicroserviceAgent) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceAgent.
func (in *MicroserviceAgent) DeepCopy() *MicroserviceAgent {
	if in == nil {
		return nil
	}
	out := new(MicroserviceAgent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceContainer) DeepCopyInto(out *MicroserviceContainer) {
	*out = *in
	if in.Commands != nil {
		in, out := &in.Commands, &out.Commands
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = new([]MicroserviceVolumeMapping)
		if **in != nil {
			in, out := *in, *out
			*out = make([]MicroserviceVolumeMapping, len(*in))
			copy(*out, *in)
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = new([]MicroserviceEnvironment)
		if **in != nil {
			in, out := *in, *out
			*out = make([]MicroserviceEnvironment, len(*in))
			copy(*out, *in)
		}
	}
	if in.ExtraHosts != nil {
		in, out := &in.ExtraHosts, &out.ExtraHosts
		*out = new([]MicroserviceExtraHost)
		if **in != nil {
			in, out := *in, *out
			*out = make([]MicroserviceExtraHost, len(*in))
			copy(*out, *in)
		}
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]MicroservicePortMapping, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceContainer.
func (in *MicroserviceContainer) DeepCopy() *MicroserviceContainer {
	if in == nil {
		return nil
	}
	out := new(MicroserviceContainer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceEnvironment) DeepCopyInto(out *MicroserviceEnvironment) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceEnvironment.
func (in *MicroserviceEnvironment) DeepCopy() *MicroserviceEnvironment {
	if in == nil {
		return nil
	}
	out := new(MicroserviceEnvironment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceExtraHost) DeepCopyInto(out *MicroserviceExtraHost) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceExtraHost.
func (in *MicroserviceExtraHost) DeepCopy() *MicroserviceExtraHost {
	if in == nil {
		return nil
	}
	out := new(MicroserviceExtraHost)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceImages) DeepCopyInto(out *MicroserviceImages) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceImages.
func (in *MicroserviceImages) DeepCopy() *MicroserviceImages {
	if in == nil {
		return nil
	}
	out := new(MicroserviceImages)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroservicePortMapping) DeepCopyInto(out *MicroservicePortMapping) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroservicePortMapping.
func (in *MicroservicePortMapping) DeepCopy() *MicroservicePortMapping {
	if in == nil {
		return nil
	}
	out := new(MicroservicePortMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroserviceVolumeMapping) DeepCopyInto(out *MicroserviceVolumeMapping) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroserviceVolumeMapping.
func (in *MicroserviceVolumeMapping) DeepCopy() *MicroserviceVolumeMapping {
	if in == nil {
		return nil
	}
	out := new(MicroserviceVolumeMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Microservices) DeepCopyInto(out *Microservices) {
	*out = *in
	if in.Microservices != nil {
		in, out := &in.Microservices, &out.Microservices
		*out = make([]Microservice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Microservices.
func (in *Microservices) DeepCopy() *Microservices {
	if in == nil {
		return nil
	}
	out := new(Microservices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TemplateVariable) DeepCopyInto(out *TemplateVariable) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TemplateVariable.
func (in *TemplateVariable) DeepCopy() *TemplateVariable {
	if in == nil {
		return nil
	}
	out := new(TemplateVariable)
	in.DeepCopyInto(out)
	return out
}
