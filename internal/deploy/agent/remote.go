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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.opt = opt

	return exe
}

func (exe *remoteExecutor) Execute() error {
	// Install the agent stack on the server
	agent := iofog.NewRemoteAgent(exe.opt.User, exe.opt.Host, exe.opt.Port, exe.opt.KeyFile, exe.opt.Name)
	err := agent.Bootstrap()
	if err != nil {
		return err
	}

	// Get Controller details
	controllers, err := config.GetControllers(exe.opt.Namespace)
	if err != nil {
		println("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}
	if len(controllers) != 1 {
		return util.NewInternalError("Only support 1 controller per namespace")
	}
	user := iofog.User{
		Name:     controllers[0].IofogUser.Name,
		Surname:  controllers[0].IofogUser.Surname,
		Email:    controllers[0].IofogUser.Email,
		Password: controllers[0].IofogUser.Password,
	}

	// Configure the agent with Controller details
	uuid, err := agent.Configure(&controllers[0], user)
	if err != nil {
		return err
	}

	// Update configuration
	configEntry := config.Agent{
		Name:    exe.opt.Name,
		User:    exe.opt.User,
		Host:    exe.opt.Host,
		KeyFile: exe.opt.KeyFile,
		UUID:    uuid,
		Created: util.NowUTC(),
	}
	err = config.UpdateAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
