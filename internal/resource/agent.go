package resource

type Agent interface {
	GetName() string
	GetUUID() string
	GetHost() string
	GetCreatedTime() string
	GetConfig() *AgentConfiguration
	GetControllerEndpoint() string
	SetName(string)
	SetUUID(string)
	SetHost(string)
	SetCreatedTime(string)
	SetConfig(*AgentConfiguration)
	Sanitize() error
	Clone() Agent
}
