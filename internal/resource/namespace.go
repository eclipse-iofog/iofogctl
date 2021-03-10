package resource

import (
	"sync"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

// Namespace is the fundamental type for managing an ECN's resources.
// Namespace's public getters will all return copies of resources.
// Namespace's API is intended to be used in parallel, hence the mutex.
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

// GetControlPlane will return a deep copy of the Namespace's ControlPlane
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

// GetControllers will return a slice of deep copied Controllers
func (ns *Namespace) GetControllers() []Controller {
	ns.mux.Lock()
	defer ns.mux.Unlock()
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
	ns.mux.Lock()
	defer ns.mux.Unlock()
	agents := ns.getAgents()
	for idx := range agents {
		if agents[idx].GetName() == name {
			return agents[idx], nil
		}
	}
	return nil, util.NewNotFoundError(name)
}

func (ns *Namespace) GetAgents() (agents []Agent) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	return ns.getAgents()
}

// getAgents will return a slice of deep copied Agents
func (ns *Namespace) getAgents() (agents []Agent) {
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
	agents := ns.getAgents()
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
	ns.RemoteAgents = []RemoteAgent{}
	ns.LocalAgents = []LocalAgent{}
}

func (ns *Namespace) Clone() *Namespace {
	ns.mux.Lock()
	defer ns.mux.Unlock()
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
	remoteAgents := make([]RemoteAgent, len(ns.RemoteAgents))
	copy(remoteAgents, ns.RemoteAgents)
	localAgents := make([]LocalAgent, len(ns.LocalAgents))
	copy(localAgents, ns.LocalAgents)
	return &Namespace{
		Name:                   ns.Name,
		KubernetesControlPlane: cpK8s,
		RemoteControlPlane:     cpRemote,
		LocalControlPlane:      cpLocal,
		LocalAgents:            localAgents,
		RemoteAgents:           remoteAgents,
		Volumes:                ns.Volumes,
	}
}

func (ns *Namespace) AddVolume(volume *Volume) error {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	if _, err := ns.getVolume(volume.Name); err == nil {
		return util.NewConflictError(ns.Name + "/" + volume.Name)
	}

	ns.Volumes = append(ns.Volumes, *volume)
	return nil
}

func (ns *Namespace) UpdateVolume(volume *Volume) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	// Replace if exists
	for idx := range ns.Volumes {
		if ns.Volumes[idx].Name == volume.Name {
			ns.Volumes[idx] = *volume
			return
		}
	}

	// Add new
	ns.Volumes = append(ns.Volumes, *volume)
}

func (ns *Namespace) DeleteVolume(name string) error {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	for idx := range ns.Volumes {
		if ns.Volumes[idx].Name == name {
			ns.Volumes = append(ns.Volumes[:idx], ns.Volumes[idx+1:]...)
			return nil
		}
	}
	return util.NewNotFoundError(ns.Name + "/" + name)
}

func (ns *Namespace) GetVolumes() (volumes []Volume) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	return append(volumes, ns.Volumes...)
}

func (ns *Namespace) GetVolume(name string) (volume Volume, err error) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	return ns.getVolume(name)
}

func (ns *Namespace) getVolume(name string) (volume Volume, err error) {
	for _, vol := range ns.Volumes {
		if vol.Name == name {
			volume = vol
			return
		}
	}

	err = util.NewNotFoundError(ns.Name + "/" + name)
	return
}
