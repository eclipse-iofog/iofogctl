package resource

type LocalAgent struct {
	Name               string              `yaml:"name,omitempty"`
	UUID               string              `yaml:"uuid,omitempty"`
	Created            string              `yaml:"created,omitempty"`
	Host               string              `yaml:"host,omitempty"`
	Package            Package             `yaml:"package,omitempty"`
	Config             *AgentConfiguration `yaml:"config,omitempty"`
	Scripts            *AgentScripts       `yaml:"scripts,omitempty"`
	ControllerEndpoint string              `yaml:"controllerEndpoint,omitempty"`
	Airgap             bool                `yaml:"airgap,omitempty"`
}

func (agent *LocalAgent) GetName() string {
	return agent.Name
}

func (agent *LocalAgent) GetUUID() string {
	return agent.UUID
}

func (agent *LocalAgent) GetHost() string {
	if agent.Config != nil && agent.Config.Host != nil && *agent.Config.Host != "" {
		return *agent.Config.Host
	}
	if agent.Host != "" {
		return agent.Host
	}
	return "localhost"
}

func (agent *LocalAgent) GetCreatedTime() string {
	return agent.Created
}

func (agent *LocalAgent) GetConfig() *AgentConfiguration {
	return agent.Config
}

func (agent *LocalAgent) GetControllerEndpoint() string {
	return agent.ControllerEndpoint
}

func (agent *LocalAgent) SetName(name string) {
	agent.Name = name
}

func (agent *LocalAgent) SetUUID(uuid string) {
	agent.UUID = uuid
}

func (agent *LocalAgent) SetHost(host string) {
	agent.Host = host
}

func (agent *LocalAgent) SetCreatedTime(time string) {
	agent.Created = time
}

func (agent *LocalAgent) SetConfig(config *AgentConfiguration) {
	agent.Config = config
}

func (agent *LocalAgent) Sanitize() error {
	if agent.Name == "" {
		agent.Name = "local"
	}
	return nil
}

func (agent *LocalAgent) Clone() Agent {
	config := agent.Config
	if agent.Config != nil {
		config = new(AgentConfiguration)
		*config = *agent.Config
	}
	scripts := agent.Scripts
	if agent.Scripts != nil {
		scripts = new(AgentScripts)
		*scripts = *agent.Scripts
	}
	return &LocalAgent{
		Name:               agent.Name,
		Host:               agent.Host,
		UUID:               agent.UUID,
		Created:            agent.Created,
		Package:            agent.Package,
		Scripts:            scripts,
		Config:             config,
		ControllerEndpoint: agent.ControllerEndpoint,
		Airgap:             agent.Airgap,
	}
}
