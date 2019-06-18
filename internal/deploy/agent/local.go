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

package deployagent

import (
	"fmt"
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	opt              *Options
	client           *iofog.LocalContainer
	localAgentConfig *iofog.LocalAgentConfig
}

func newLocalExecutor(opt *Options, client *iofog.LocalContainer) *localExecutor {
	return &localExecutor{
		opt:              opt,
		client:           client,
		localAgentConfig: iofog.NewLocalAgentConfig(opt.Name, opt.Image),
	}
}

func (exe *localExecutor) provisionAgent() (string, error) {
	// Get agent
	agent := iofog.NewLocalAgent(exe.localAgentConfig, exe.client)

	// Get Controller details
	controllers, err := config.GetControllers(exe.opt.Namespace)
	if err != nil {
		println("You must deploy a Controller to a namespace before deploying any Agents")
		return "", err
	}
	if len(controllers) != 1 {
		return "", util.NewInternalError("Only support 1 controller per namespace")
	}
	user := iofog.User{
		Name:     controllers[0].IofogUser.Name,
		Surname:  controllers[0].IofogUser.Surname,
		Email:    controllers[0].IofogUser.Email,
		Password: controllers[0].IofogUser.Password,
	}

	// Configure the agent with Controller details
	return agent.Configure(&controllers[0], user)
}

func (exe *localExecutor) Execute() error {
	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy agent image
	if exe.opt.Image == "" {
		exe.opt.Image = exe.localAgentConfig.DefaultImage
	}

	if _, err = exe.client.DeployContainer(&exe.localAgentConfig.LocalContainerConfig); err != nil {
		return err
	}

	agentContainerName := exe.localAgentConfig.ContainerName

	// Wait for agent
	if err = exe.client.WaitForCommand(
		regexp.MustCompile("401 Unauthorized"),
		"curl",
		"--request",
		"GET",
		"--url",
		fmt.Sprintf("http://%s:%s/v2/status", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host),
	); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	// Provision agent
	uuid, err := exe.provisionAgent()
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	// Update configuration
	agentIP := fmt.Sprintf("%s:%s", exe.localAgentConfig.Host, exe.localAgentConfig.Ports[0].Host)
	configEntry := config.Agent{
		Name: exe.opt.Name,
		User: currUser.Username,
		Host: agentIP,
		UUID: uuid,
	}
	err = config.AddAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	if err = config.Flush(); err != nil {
		if cleanErr := exe.client.CleanContainer(agentContainerName); cleanErr != nil {
			fmt.Printf("Could not clean container %s\n", agentContainerName)
		}
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
