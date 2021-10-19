/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package resource

type LocalAgent struct {
	Name               string              `yaml:"name,omitempty"`
	UUID               string              `yaml:"uuid,omitempty"`
	Container          Container           `yaml:"container,omitempty"`
	Created            string              `yaml:"created,omitempty"`
	Host               string              `yaml:"host,omitempty"`
	Config             *AgentConfiguration `yaml:"config,omitempty"`
	ControllerEndpoint string              `yaml:"controllerEndpoint,omitempty"`
}

func (agent *LocalAgent) GetName() string {
	return agent.Name
}

func (agent *LocalAgent) GetUUID() string {
	return agent.UUID
}

func (agent *LocalAgent) GetHost() string {
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
	return &LocalAgent{
		Name:               agent.Name,
		Host:               agent.Host,
		UUID:               agent.UUID,
		Created:            agent.Created,
		Container:          agent.Container,
		Config:             config,
		ControllerEndpoint: agent.ControllerEndpoint,
	}
}
