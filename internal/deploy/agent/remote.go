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
	"strconv"

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

//
// Install iofog-agent stack on an agent host
//
func (exe *remoteExecutor) Execute() error {

	// Get Controllers from namespace
	controllers, err := config.GetControllers(exe.opt.Namespace)

	// Do we actually have any controllers?
	if err != nil {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}

	// Did we have more than one controller?
	if len(controllers) != 1 {
		return util.NewInternalError("Only support 1 controller per namespace")
	}

	// Create our user object
	user := iofog.User{
		Name:     controllers[0].IofogUser.Name,
		Surname:  controllers[0].IofogUser.Surname,
		Email:    controllers[0].IofogUser.Email,
		Password: controllers[0].IofogUser.Password,
	}

	// Try and get our ssh connection
	util.PrintInfo("Attempting to connect to Agent [" + exe.opt.Name + "] as '" +
		exe.opt.User + "@" + exe.opt.Host + ":" + strconv.Itoa(exe.opt.Port))

	agent := iofog.NewRemoteAgent(exe.opt.User, exe.opt.Host, exe.opt.Port, exe.opt.KeyFile, exe.opt.Name)

	// Try the install
	agentError := agent.Bootstrap()
	if agentError != nil {
		return err
	}

	util.PrintInfo("Agent install successful. Provisioning to Controller.")

	// Configure the agent with Controller details
	uuid, err := agent.Configure(&controllers[0], user)
	if err != nil {
		return err
	}

	util.PrintInfo("Agent install successful. Connecting with Controller.")

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
