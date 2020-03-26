package resource

import "github.com/eclipse-iofog/iofogctl/v2/pkg/util"

type Namespace struct {
	Name                   string                  `yaml:"name,omitempty"`
	KubernetesControlPlane *KubernetesControlPlane `yaml:"k8sControlPlane,omitempty"`
	RemoteControlPlane     *RemoteControlPlane     `yaml:"remoteControlPlane,omitempty"`
	LocalControlPlane      *LocalControlPlane      `yaml:"localControlPlane,omitempty"`
	LocalAgents            []LocalAgent            `yaml:"localAgents,omitempty"`
	RemoteAgents           []RemoteAgent           `yaml:"remoteAgents,omitempty"`
	Volumes                []Volume                `yaml:"volumes,omitempty"`
	Created                string                  `yaml:"created,omitempty"`
}

func (ns *Namespace) GetControlPlane() (ControlPlane, error) {
	if ns.KubernetesControlPlane != nil {
		return ns.KubernetesControlPlane, nil
	}
	if ns.RemoteControlPlane != nil {
		return ns.RemoteControlPlane, nil
	}
	if ns.LocalControlPlane != nil {
		return ns.LocalControlPlane, nil
	}
	return nil, NewNoControlPlaneError(ns.Name)
}

func (ns *Namespace) SetControlPlane(baseControlPlane ControlPlane) {
	switch controlPlane := baseControlPlane.(type) {
	case *KubernetesControlPlane:
		ns.KubernetesControlPlane = controlPlane
	case *RemoteControlPlane:
		ns.RemoteControlPlane = controlPlane
	case *LocalControlPlane:
		ns.LocalControlPlane = controlPlane
	}
}

func (ns *Namespace) DeleteControlPlane() {
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
		agents = append(agents, &ns.RemoteAgents[idx])
	}
	// Local
	for idx := range ns.LocalAgents {
		agents = append(agents, &ns.LocalAgents[idx])
	}
	return
}

func (ns *Namespace) UpdateAgent(baseAgent Agent) error {
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
