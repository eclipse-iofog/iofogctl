package resource

import "github.com/eclipse-iofog/iofogctl/v2/pkg/util"

type Namespace struct {
	Name                   string                  `yaml:"name,omitempty"`
	kubernetesControlPlane *KubernetesControlPlane `yaml:"k8sControlPlane,omitempty"`
	remoteControlPlane     *RemoteControlPlane     `yaml:"remoteControlPlane,omitempty"`
	localControlPlane      *LocalControlPlane      `yaml:"localControlPlane,omitempty"`
	localAgents            []LocalAgent            `yaml:"localAgents,omitempty"`
	remoteAgents           []RemoteAgent           `yaml:"remoteAgents,omitempty"`
	Volumes                []Volume                `yaml:"volumes,omitempty"`
	Created                string                  `yaml:"created,omitempty"`
}

func (ns *Namespace) GetControlPlane() (ControlPlane, error) {
	if ns.kubernetesControlPlane != nil {
		return ns.kubernetesControlPlane.Clone(), nil
	}
	if ns.remoteControlPlane != nil {
		return ns.remoteControlPlane.Clone(), nil
	}
	if ns.localControlPlane != nil {
		return ns.localControlPlane.Clone(), nil
	}
	return nil, NewNoControlPlaneError(ns.Name)
}

func (ns *Namespace) SetControlPlane(baseControlPlane ControlPlane) {
	switch controlPlane := baseControlPlane.(type) {
	case *KubernetesControlPlane:
		ns.kubernetesControlPlane = controlPlane
		ns.remoteControlPlane = nil
		ns.localControlPlane = nil
	case *RemoteControlPlane:
		ns.remoteControlPlane = controlPlane
		ns.kubernetesControlPlane = nil
		ns.localControlPlane = nil
	case *LocalControlPlane:
		ns.localControlPlane = controlPlane
		ns.kubernetesControlPlane = nil
		ns.remoteControlPlane = nil
	}
}

func (ns *Namespace) DeleteControlPlane() {
	ns.kubernetesControlPlane = nil
	ns.remoteControlPlane = nil
	ns.localControlPlane = nil
}

func (ns *Namespace) GetControllers() []Controller {
	if ns.kubernetesControlPlane != nil {
		return ns.kubernetesControlPlane.GetControllers()
	}
	if ns.remoteControlPlane != nil {
		return ns.remoteControlPlane.GetControllers()
	}
	if ns.localControlPlane != nil {
		return ns.localControlPlane.GetControllers()
	}
	return []Controller{}
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
	for idx := range ns.remoteAgents {
		agents = append(agents, ns.remoteAgents[idx].Clone())
	}
	// Local
	for idx := range ns.localAgents {
		agents = append(agents, ns.localAgents[idx].Clone())
	}
	return
}

func (ns *Namespace) UpdateAgent(baseAgent Agent) error {
	switch agent := baseAgent.(type) {
	case *LocalAgent:
		for idx := range ns.localAgents {
			if ns.localAgents[idx].GetName() == baseAgent.GetName() {
				agent, ok := baseAgent.(*LocalAgent)
				if !ok {
					return util.NewError("Could not convert Agent to Local Agent during update")
				}

				ns.localAgents[idx] = *agent
				return nil
			}
		}
		ns.localAgents = append(ns.localAgents, *agent)
		return nil
	case *RemoteAgent:
		for idx := range ns.remoteAgents {
			if ns.remoteAgents[idx].GetName() == baseAgent.GetName() {
				agent, ok := baseAgent.(*RemoteAgent)
				if !ok {
					return util.NewError("Could not convert Agent to Remote Agent during update")
				}

				ns.remoteAgents[idx] = *agent
				return nil
			}
		}
		ns.remoteAgents = append(ns.remoteAgents, *agent)
		return nil
	}

	return nil
}

func (ns *Namespace) AddAgent(baseAgent Agent) error {
	agents := ns.GetAgents()
	for idx := range agents {
		if agents[idx].GetName() == baseAgent.GetName() {
			return util.NewConflictError(baseAgent.GetName())
		}
	}
	switch agent := baseAgent.(type) {
	case *LocalAgent:
		ns.localAgents = append(ns.localAgents, *agent)
	case *RemoteAgent:
		ns.remoteAgents = append(ns.remoteAgents, *agent)
	}
	return nil
}

func (ns *Namespace) DeleteAgent(name string) error {
	for idx := range ns.localAgents {
		if ns.localAgents[idx].Name == name {
			ns.localAgents = append(ns.localAgents[:idx], ns.localAgents[idx+1:]...)
			return nil
		}
	}
	for idx := range ns.remoteAgents {
		if ns.remoteAgents[idx].Name == name {
			ns.remoteAgents = append(ns.remoteAgents[:idx], ns.remoteAgents[idx+1:]...)
			return nil
		}
	}
	return util.NewNotFoundError(name)
}

func (ns *Namespace) DeleteAgents() {
	ns.remoteAgents = make([]RemoteAgent, 0)
	ns.localAgents = make([]LocalAgent, 0)
}
