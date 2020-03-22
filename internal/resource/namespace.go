package resource

type Namespace struct {
	Name                   string                  `yaml:"name,omitempty"`
	KubernetesControlPlane *KubernetesControlPlane `yaml:"k8sControlPlane,omitempty"`
	RemoteControlPlane     *RemoteControlPlane     `yaml:"remoteControlPlane,omitempty"`
	LocalControlPlane      *LocalControlPlane      `yaml:"localControlPlane,omitempty"`
	Agents                 []Agent                 `yaml:"agents,omitempty"`
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
