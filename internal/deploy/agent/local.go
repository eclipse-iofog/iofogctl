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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	opt    *Options
	client *iofog.LocalContainer
}

func newLocalExecutor(opt *Options, client *iofog.LocalContainer) *localExecutor {
	return &localExecutor{
		opt:    opt,
		client: client,
	}
}

func (exe *localExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy agent image
	if exe.opt.Image == "" {
		return util.NewInputError("No agent image specified")
	}
	agentPortMap := make(map[string]*iofog.LocalContainerPort)
	agentPortMap["54321"] = &iofog.LocalContainerPort{
		Protocol: "tcp",
		Port:     "54321",
	} // 54321:54321/tcp
	err = exe.client.DeployContainer(exe.opt.Image, fmt.Sprintf("iofog-agent-%s", exe.opt.Name), agentPortMap)
	if err != nil {
		return err
	}

	// Update configuration
	configEntry := config.Agent{
		Name: exe.opt.Name,
		User: currUser.Username,
		Host: "0.0.0.0:54321",
	}
	err = config.AddAgent(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
