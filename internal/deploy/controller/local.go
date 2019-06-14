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

package deploycontroller

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

func (exe *localExecutor) deployContainers() error {
	// Deploy controller image
	controllerImg, exists := exe.opt.Images["controller"]
	if !exists {
		return util.NewInputError("No controller image specified")
	}
	controllerPortMap := make(map[string]*iofog.LocalContainerPort)
	controllerPortMap["51121"] = &iofog.LocalContainerPort{
		Protocol: "tcp",
		Port:     "51121",
	} // 51121:51121/tcp
	err := exe.client.DeployContainer(controllerImg, "iofog-controller", controllerPortMap)
	if err != nil {
		return err
	}

	// Deploy controller image
	connectorImg, exists := exe.opt.Images["connector"]
	if !exists {
		return util.NewInputError("No connector image specified")
	}
	connectorPortMap := make(map[string]*iofog.LocalContainerPort)
	connectorPortMap["53321"] = &iofog.LocalContainerPort{
		Protocol: "tcp",
		Port:     "8080",
	} // 53321:8080/tcp
	return exe.client.DeployContainer(connectorImg, "iofog-connector", connectorPortMap)
}

func (exe *localExecutor) Execute() error {
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	err = exe.deployContainers()
	if err != nil {
		return err
	}

	// TODO - SET UP

	// Update configuration
	configEntry := config.Controller{
		Name:   exe.opt.Name,
		User:   currUser.Username,
		Host:   "0.0.0.0:51121",
		Images: exe.opt.Images,
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
