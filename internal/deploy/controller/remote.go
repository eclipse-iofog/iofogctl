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
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace    string
	ctrl         *config.Controller
	controlPlane config.ControlPlane
}

func newRemoteExecutor(namespace string, ctrl *config.Controller, controlPlane config.ControlPlane) *remoteExecutor {
	d := &remoteExecutor{}
	d.namespace = namespace
	d.ctrl = ctrl
	d.controlPlane = controlPlane
	return d
}

func (exe *remoteExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *remoteExecutor) Execute() (err error) {
	// Instantiate deployer
	controllerOptions := &install.ControllerOptions{
		User:            exe.ctrl.SSH.User,
		Host:            exe.ctrl.Host,
		Port:            exe.ctrl.SSH.Port,
		PrivKeyFilename: exe.ctrl.SSH.KeyFile,
		Version:         exe.ctrl.Package.Version,
		Repo:            exe.ctrl.Package.Repo,
		Token:           exe.ctrl.Package.Token,
	}
	deployer := install.NewController(controllerOptions)

	// Set database configuration
	if exe.controlPlane.Database.Host != "" {
		db := exe.controlPlane.Database
		deployer.SetControllerExternalDatabase(db.Host, db.User, db.Password, db.Provider, db.DatabaseName, db.Port)
	}

	// Deploy Controller
	if err = deployer.Install(); err != nil {
		return
	}
	// Update controller (its a pointer, this is returned to caller)
	exe.ctrl.Endpoint = exe.ctrl.Host + ":" + iofog.ControllerPortString

	// Deploy system agent to host internal router
	install.Verbose("Deploying system agent")
	agentConfig := config.Agent{
		Name: iofog.VanillaRouterAgentName,
		Host: exe.ctrl.Host,
		SSH:  exe.ctrl.SSH,
	}
	agentDeployGenericExecutor, err := deployagent.NewDeployExecutor(exe.namespace, &agentConfig)
	// Convert executor to be able to setConfig
	agentDeployExecutor, ok := agentDeployGenericExecutor.(deployagent.AgentExecutor)
	if !ok {
		return util.NewInternalError("Could not convert executor")
	}
	// Configure agent to be system agent
	isSystem := true
	deployAgentConfig := config.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			IsSystem: &isSystem,
		},
	}
	agentDeployExecutor.SetAgentConfig(&deployAgentConfig)
	if err != nil {
		return err
	}

	if err = agentDeployExecutor.Execute(); err != nil {
		return err
	}

	return
}
