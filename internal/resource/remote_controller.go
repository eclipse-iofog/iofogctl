package resource

import (
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type SystemAgentConfig struct {
	Package            Package             `yaml:"package,omitempty"`
	Scripts            *AgentScripts       `yaml:"scripts,omitempty"` // Custom scripts
	AgentConfiguration *AgentConfiguration `yaml:"config,omitempty"`  // Configurable config
}

type RemoteController struct {
	Name        string             `yaml:"name"`
	Host        string             `yaml:"host"`
	SSH         SSH                `yaml:"ssh,omitempty"`
	TLS         *ControlPlaneTLS   `yaml:"tls,omitempty"`
	SystemAgent *SystemAgentConfig `yaml:"systemAgent,omitempty"`
	Endpoint    string             `yaml:"endpoint,omitempty"`
	Created     string             `yaml:"created,omitempty"`
	Airgap      bool               `yaml:"airgap,omitempty"`
}

func (ctrl *RemoteController) GetName() string {
	return ctrl.Name
}

func (ctrl *RemoteController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl *RemoteController) GetCreatedTime() string {
	return ctrl.Created
}

func (ctrl *RemoteController) SetName(name string) {
	ctrl.Name = name
}

func (ctrl *RemoteController) Sanitize() (err error) {
	if ctrl.Host != "" && ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}
	return
}

func (ctrl *RemoteController) Clone() Controller {
	var systemAgent *SystemAgentConfig
	if ctrl.SystemAgent != nil {
		systemAgent = new(SystemAgentConfig)
		*systemAgent = *ctrl.SystemAgent
	}
	return &RemoteController{
		Name:        ctrl.Name,
		Host:        ctrl.Host,
		SSH:         ctrl.SSH,
		TLS:         ctrl.TLS,
		SystemAgent: systemAgent,
		Endpoint:    ctrl.Endpoint,
		Created:     ctrl.Created,
		Airgap:      ctrl.Airgap,
	}
}

func (ctrl *RemoteController) ValidateSSH() error {
	if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.Port == 0 || ctrl.SSH.KeyFile == "" {
		return NewNoSSHConfigError("Controller")
	}
	return nil
}
