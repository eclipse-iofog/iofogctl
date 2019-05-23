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

// Namespace export
type Namespace struct {
	Name string
}

type namespace struct {
	Name       string     `mapstructure:"name"`
	Controllers []Controller `mapstructure:"controller"`
	Agents     []Agent    `mapstructure:"agents"`
}

type configuration struct {
	Namespaces []namespace `mapstructure:"namespaces"`
}