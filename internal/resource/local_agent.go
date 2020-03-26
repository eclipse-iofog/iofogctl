/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
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
	Name      string    `yaml:"name,omitempty"`
	UUID      string    `yaml:"uuid,omitempty"`
	Container Container `yaml:"container,omitempty"`
	Created   string    `yaml:"created,omitempty"`
	Host      string    `yaml:"host,omitempty"`
}

func (agent LocalAgent) GetName() string {
	return agent.Name
}

func (agent LocalAgent) GetUUID() string {
	return agent.UUID
}

func (agent LocalAgent) GetHost() string {
	return "localhost"
}

func (agent LocalAgent) GetCreatedTime() string {
	return agent.Created
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

func (agent *LocalAgent) Sanitize() error {
	return nil
}
