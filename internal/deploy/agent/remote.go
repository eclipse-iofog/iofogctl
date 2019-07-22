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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	opt  *Options
	uuid string
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.opt = opt

	return exe
}

//
// Install iofog-agent stack on an agent host
//
func (exe *remoteExecutor) Execute() error {

	configEntry, err := DeployAgent(exe.opt)
	if err != nil {
		return err
	}

	if err = config.UpdateAgent(exe.opt.Namespace, configEntry); err != nil {
		return err
	}

	return config.Flush()
}

func DeployAgent(opt *Options) (configEntry config.Agent, err error) {
	// Get Controllers from namespace
	controllers, err := config.GetControllers(opt.Namespace)

	// Do we actually have any controllers?
	if err != nil {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Did we have more than one controller?
	if len(controllers) != 1 {
		err = util.NewInternalError("Only support 1 controller per namespace")
		return
	}

	// Connect to agent via SSH
	agent := install.NewRemoteAgent(opt.User, opt.Host, opt.Port, opt.KeyFile, opt.Name)

	// Try the install
	err = agent.Bootstrap()
	if err != nil {
		return
	}

	// Create our user object
	user := client.User{
		Name:     controllers[0].IofogUser.Name,
		Surname:  controllers[0].IofogUser.Surname,
		Email:    controllers[0].IofogUser.Email,
		Password: controllers[0].IofogUser.Password,
	}

	// Configure the agent with Controller details
	uuid, err := agent.Configure(&controllers[0], user)
	if err != nil {
		return
	}

	configEntry = config.Agent{
		Name:    opt.Name,
		User:    opt.User,
		Host:    opt.Host,
		KeyFile: opt.KeyFile,
		UUID:    uuid,
		Created: util.NowUTC(),
	}
	return
}
