package resource

import (
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"sync"
)

type Namespace struct {
	Name                   string                  `yaml:"name,omitempty"`
	KubernetesControlPlane *KubernetesControlPlane `yaml:"k8sControlPlane,omitempty"`
	RemoteControlPlane     *RemoteControlPlane     `yaml:"remoteControlPlane,omitempty"`
	LocalControlPlane      *LocalControlPlane      `yaml:"localControlPlane,omitempty"`
	LocalAgents            []LocalAgent            `yaml:"localAgents,omitempty"`
	RemoteAgents           []RemoteAgent           `yaml:"remoteAgents,omitempty"`
	Volumes                []Volume                `yaml:"volumes,omitempty"`
	Created                string                  `yaml:"created,omitempty"`
	mux                    sync.Mutex
}

func (ns *Namespace) GetControlPlane() (ControlPlane, error) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	if ns.KubernetesControlPlane != nil {
		return ns.KubernetesControlPlane.Clone(), nil
	}
	if ns.RemoteControlPlane != nil {
		return ns.RemoteControlPlane.Clone(), nil
	}
	if ns.LocalControlPlane != nil {
		return ns.LocalControlPlane.Clone(), nil
	}
	return nil, NewNoControlPlaneError(ns.Name)
}

func (ns *Namespace) SetControlPlane(baseControlPlane ControlPlane) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	switch controlPlane := baseControlPlane.(type) {
	case *KubernetesControlPlane:
		ns.KubernetesControlPlane = controlPlane
		ns.RemoteControlPlane = nil
		ns.LocalControlPlane = nil
	case *RemoteControlPlane:
		ns.RemoteControlPlane = controlPlane
		ns.KubernetesControlPlane = nil
		ns.LocalControlPlane = nil
	case *LocalControlPlane:
		ns.LocalControlPlane = controlPlane
		ns.KubernetesControlPlane = nil
		ns.RemoteControlPlane = nil
	}
}

func (ns *Namespace) DeleteControlPlane() {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	ns.KubernetesControlPlane = nil
	ns.RemoteControlPlane = nil
	ns.LocalControlPlane = nil
}

func (ns *Namespace) GetControllers() []Controller {
	if ns.KubernetesControlPlane != nil {
		return ns.KubernetesControlPlane.GetControllers()
	}
	if ns.RemoteControlPlane != nil {
		return ns.RemoteControlPlane.GetControllers()
	}
	if ns.LocalControlPlane != nil {
		return ns.LocalControlPlane.GetControllers()
	}
	return []Controller{}
}

func (ns *Namespace) DeleteController(name string) (err error) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	if ns.KubernetesControlPlane != nil {
		return ns.KubernetesControlPlane.DeleteController(name)
	}
	if ns.RemoteControlPlane != nil {
		return ns.RemoteControlPlane.DeleteController(name)
	}
	if ns.LocalControlPlane != nil {
		return ns.LocalControlPlane.DeleteController(name)
	}
	return
}

func (ns *Namespace) GetAgent(name string) (Agent, error) {
	agents := ns.GetAgents()
	for idx := range agents {
		if agents[idx].GetName() == name {
			return agents[idx], nil
		}
	}
	return nil, util.NewError("Could not find Agent " + name)
}

func (ns *Namespace) GetAgents() (agents []Agent) {
	// K8s / Remote
	for idx := range ns.RemoteAgents {
		agents = append(agents, ns.RemoteAgents[idx].Clone())
	}
	// Local
	for idx := range ns.LocalAgents {
		agents = append(agents, ns.LocalAgents[idx].Clone())
	}
	return
}

func (ns *Namespace) UpdateAgent(baseAgent Agent) error {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	switch agent := baseAgent.(type) {
	case *LocalAgent:
		for idx := range ns.LocalAgents {
			if ns.LocalAgents[idx].GetName() == baseAgent.GetName() {
				agent, ok := baseAgent.(*LocalAgent)
				if !ok {
					return util.NewError("Could not convert Agent to Local Agent during update")
				}

				ns.LocalAgents[idx] = *agent
				return nil
			}
		}
		ns.LocalAgents = append(ns.LocalAgents, *agent)
		return nil
	case *RemoteAgent:
		for idx := range ns.RemoteAgents {
			if ns.RemoteAgents[idx].GetName() == baseAgent.GetName() {
				agent, ok := baseAgent.(*RemoteAgent)
				if !ok {
					return util.NewError("Could not convert Agent to Remote Agent during update")
				}

				ns.RemoteAgents[idx] = *agent
				return nil
			}
		}
		ns.RemoteAgents = append(ns.RemoteAgents, *agent)
		return nil
	}

	return nil
}

func (ns *Namespace) AddAgent(baseAgent Agent) error {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	agents := ns.GetAgents()
	for idx := range agents {
		if agents[idx].GetName() == baseAgent.GetName() {
			return util.NewConflictError(baseAgent.GetName())
		}
	}
	switch agent := baseAgent.(type) {
	case *LocalAgent:
		ns.LocalAgents = append(ns.LocalAgents, *agent)
	case *RemoteAgent:
		ns.RemoteAgents = append(ns.RemoteAgents, *agent)
	}
	return nil
}

func (ns *Namespace) DeleteAgent(name string) error {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	for idx := range ns.LocalAgents {
		if ns.LocalAgents[idx].Name == name {
			ns.LocalAgents = append(ns.LocalAgents[:idx], ns.LocalAgents[idx+1:]...)
			return nil
		}
	}
	for idx := range ns.RemoteAgents {
		if ns.RemoteAgents[idx].Name == name {
			ns.RemoteAgents = append(ns.RemoteAgents[:idx], ns.RemoteAgents[idx+1:]...)
			return nil
		}
	}
	return util.NewNotFoundError(name)
}

func (ns *Namespace) DeleteAgents() {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	ns.RemoteAgents = make([]RemoteAgent, 0)
	ns.LocalAgents = make([]LocalAgent, 0)
}

func (ns *Namespace) Clone() *Namespace {
	var cpK8s *KubernetesControlPlane
	var cpRemote *RemoteControlPlane
	var cpLocal *LocalControlPlane
	if ns.KubernetesControlPlane != nil {
		cpK8s = ns.KubernetesControlPlane.Clone().(*KubernetesControlPlane)
	}
	if ns.RemoteControlPlane != nil {
		cpRemote = ns.RemoteControlPlane.Clone().(*RemoteControlPlane)
	}
	if ns.LocalControlPlane != nil {
		cpLocal = ns.LocalControlPlane.Clone().(*LocalControlPlane)
	}
	agentsRemote := make([]RemoteAgent, 0)
	for idx := range ns.RemoteAgents {
		agentsRemote = append(agentsRemote, ns.RemoteAgents[idx])
	}
	agentsLocal := make([]LocalAgent, 0)
	for idx := range ns.LocalAgents {
		agentsLocal = append(agentsLocal, ns.LocalAgents[idx])
	}
	return &Namespace{
		Name:                   ns.Name,
		KubernetesControlPlane: cpK8s,
		RemoteControlPlane:     cpRemote,
		LocalControlPlane:      cpLocal,
		LocalAgents:            agentsLocal,
		RemoteAgents:           agentsRemote,
		Volumes:                ns.Volumes,
	}
}

func (ns *Namespace) AddVolume(volume Volume) error {
	if _, err := ns.GetVolume(volume.Name); err == nil {
		return util.NewConflictError(ns.Name + "/" + volume.Name)
	}

	ns.Volumes = append(ns.Volumes, volume)
	return nil
}

func (ns *Namespace) UpdateVolume(volume Volume) {
	// Replace if exists
	for idx := range ns.Volumes {
		if ns.Volumes[idx].Name == volume.Name {
			ns.Volumes[idx] = volume
			return
		}
	}

	// Add new
	ns.Volumes = append(ns.Volumes, volume)
	return
}

func (ns *Namespace) DeleteVolume(name string) error {
	for idx := range ns.Volumes {
		if ns.Volumes[idx].Name == name {
			ns.Volumes = append(ns.Volumes[:idx], ns.Volumes[idx+1:]...)
			return nil
		}
	}
	return util.NewNotFoundError(ns.Name + "/" + name)
}

func (ns *Namespace) GetVolumes() []Volume {
	return ns.Volumes
}

func (ns *Namespace) GetVolume(name string) (agent Volume, err error) {
	for _, ag := range ns.Volumes {
		if ag.Name == name {
			agent = ag
			return
		}
	}

	err = util.NewNotFoundError(ns.Name + "/" + name)
	return
}
