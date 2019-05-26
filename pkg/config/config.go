package config

import (
)

// Controller export
type Controller struct {
	Name string `mapstructure:"name"`
	User string `mapstructure:"user"`
}

// Agent export
type Agent struct {
	Name string `mapstructure:"name"`
	User string `mapstructure:"user"`
}

// Microservice export
type Microservice struct {
	Name string `mapstructure:"name"`
	Flow string `mapstructure:"flow"`
}

// Namespace export
type Namespace struct {
	Name string
}

type namespace struct {
	Name       string     `mapstructure:"name"`
	Controllers []Controller `mapstructure:"controllers"`
	Agents     []Agent    `mapstructure:"agents"`
	Microservices []Microservice `mapstructure:"microservices"`
}

type configuration struct {
	Namespaces []namespace `mapstructure:"namespaces"`
}