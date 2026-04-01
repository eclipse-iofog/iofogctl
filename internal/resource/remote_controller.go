/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
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

import (
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type ControllerScripts struct {
	install.ControllerProcedures `yaml:",inline"`
	Directory                    string `yaml:"dir"` // Location of scripts
}

type SystemAgentConfig struct {
	Package            Package             `yaml:"package,omitempty"`
	Scripts            *AgentScripts       `yaml:"scripts,omitempty"` // Custom scripts
	AgentConfiguration *AgentConfiguration `yaml:"config,omitempty"`  // Configurable config
}

type RemoteController struct {
	RemoteControllerConfig `yaml:",inline"`
	Name                   string             `yaml:"name"`
	Host                   string             `yaml:"host"`
	SSH                    SSH                `yaml:"ssh,omitempty"`
	Endpoint               string             `yaml:"endpoint,omitempty"`
	Created                string             `yaml:"created,omitempty"`
	Scripts                *ControllerScripts `yaml:"scripts,omitempty"`
	SystemAgent            *SystemAgentConfig `yaml:"systemAgent,omitempty"` // Per-controller system agent config
	Airgap                 bool               `yaml:"airgap,omitempty"`
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
	// Fix SSH port
	if ctrl.Host != "" && ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	// Format file paths
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}
	return
}

func (ctrl *RemoteController) Clone() Controller {
	scripts := ctrl.Scripts
	if ctrl.Scripts != nil {
		scripts = new(ControllerScripts)
		*scripts = *ctrl.Scripts
	}
	systemAgent := ctrl.SystemAgent
	if ctrl.SystemAgent != nil {
		systemAgent = new(SystemAgentConfig)
		*systemAgent = *ctrl.SystemAgent
	}
	return &RemoteController{
		RemoteControllerConfig: ctrl.RemoteControllerConfig,
		Name:                   ctrl.Name,
		Host:                   ctrl.Host,
		SSH:                    ctrl.SSH,
		Endpoint:               ctrl.Endpoint,
		Created:                ctrl.Created,
		Scripts:                scripts,
		SystemAgent:            systemAgent,
		Airgap:                 ctrl.Airgap,
	}
}

func (ctrl *RemoteController) ValidateSSH() error {
	if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.Port == 0 || ctrl.SSH.KeyFile == "" {
		return NewNoSSHConfigError("Controller")
	}
	return nil
}
