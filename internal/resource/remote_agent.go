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

import "github.com/eclipse-iofog/iofogctl/v2/pkg/util"

type RemoteAgent struct {
	Name    string  `yaml:"name"`
	Host    string  `yaml:"host"`
	SSH     SSH     `yaml:"ssh"`
	UUID    string  `yaml:"uuid,omitempty"`
	Created string  `yaml:"created,omitempty"`
	Package Package `yaml:"package,omitempty"`
}

func (agent RemoteAgent) GetName() string {
	return agent.Name
}

func (agent RemoteAgent) GetUUID() string {
	return agent.UUID
}

func (agent RemoteAgent) GetHost() string {
	return agent.Host
}

func (agent RemoteAgent) GetCreatedTime() string {
	return agent.Created
}

func (agent *RemoteAgent) SetName(name string) {
	agent.Name = name
}

func (agent *RemoteAgent) SetUUID(uuid string) {
	agent.UUID = uuid
}

func (agent *RemoteAgent) SetHost(host string) {
	agent.Host = host
}

func (agent *RemoteAgent) SetCreatedTime(time string) {
	agent.Created = time
}

func (agent *RemoteAgent) Sanitize() (err error) {
	if agent.SSH.Port == 0 {
		agent.SSH.Port = 22
	}
	if agent.SSH.KeyFile, err = util.FormatPath(agent.SSH.KeyFile); err != nil {
		return
	}
	return
}

func (agent *RemoteAgent) ValidateSSH() error {
	if agent.Host == "" || agent.SSH.User == "" || agent.SSH.Port == 0 || agent.SSH.KeyFile == "" {
		return NewNoSSHConfigError("Agent")
	}
	return nil
}
