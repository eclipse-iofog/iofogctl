package config

// Controller export
type Controller struct {
	Name       string `mapstructure:"name"`
	User       string `mapstructure:"user"`
	Host       string `mapstructure:"host"`
	KeyFile    string `mapstructure:"keyFile"`
	KubeConfig string `mapstructure:"kubeConfig"`
}

// Agent export
type Agent struct {
	Name    string `mapstructure:"name"`
	User    string `mapstructure:"user"`
	Host    string `mapstructure:"host"`
	KeyFile string `mapstructure:"keyFile"`
}

// Microservice export
type Microservice struct {
	Name string `mapstructure:"name"`
	Flow string `mapstructure:"flow"`
}

type Namespace struct {
	Name          string         `mapstructure:"name"`
	Controllers   []Controller   `mapstructure:"controllers"`
	Agents        []Agent        `mapstructure:"agents"`
	Microservices []Microservice `mapstructure:"microservices"`
}

type configuration struct {
	Namespaces []Namespace `mapstructure:"namespaces"`
}
