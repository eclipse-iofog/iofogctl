package config

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string
	Surname  string
	Email    string
	Password string
}

// Controller contains information for configuring a controller
type Controller struct {
	Name       string            `mapstructure:"name"`
	User       string            `mapstructure:"user"`
	Host       string            `mapstructure:"host"`
	KeyFile    string            `mapstructure:"keyFile"`
	KubeConfig string            `mapstructure:"kubeConfig"`
	Endpoint   string            `mapstructure:"endpoint"`
	IofogUser  IofogUser         `mapstructure:"iofogUser"`
	Created    string            `mapstructure:"created"`
	Images     map[string]string `mapstructure:"images"`
}

// Agent contains information for configuring an agent
type Agent struct {
	Name    string `mapstructure:"name"`
	User    string `mapstructure:"user"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	KeyFile string `mapstructure:"keyFile"`
	UUID    string `mapstructure:"uuid"`
	Created string `mapstructure:"created"`
}

// Microservice contains information for configuring a microservice
type Microservice struct {
	Name    string `mapstructure:"name"`
	Flow    string `mapstructure:"flow"`
	Created string `mapstructure:"created"`
}

// Namespace contains information for configuring a namespace
type Namespace struct {
	Name          string         `mapstructure:"name"`
	Controllers   []Controller   `mapstructure:"controllers"`
	Agents        []Agent        `mapstructure:"agents"`
	Microservices []Microservice `mapstructure:"microservices"`
	Created       string         `mapstructure:"created"`
}

// configuration contains the unmarshalled configuration file
type configuration struct {
	Namespaces []Namespace `mapstructure:"namespaces"`
}
