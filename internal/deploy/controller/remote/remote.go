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

package deployremotecontroller

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	namespace    string
	controlPlane *rsc.RemoteControlPlane
	controller   *rsc.RemoteController
}

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshall file
	controller, err := UnmarshallYAML(opt.Yaml)
	if err != nil {
		return
	}

	if len(opt.Name) > 0 {
		controller.Name = opt.Name
	}

	// Validate
	if err = Validate(controller); err != nil {
		return
	}

	// Get the Control Plane
	baseControlPlane, err := config.GetControlPlane(opt.Namespace)
	if err != nil {
		return
	}
	controlPlane, ok := baseControlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		err = util.NewError("Could not convert Control Plane to Remote Control Plane")
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, controlPlane, controller)
}

func newExecutor(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) *remoteExecutor {
	return &remoteExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
		controller:   controller,
	}
}

func (exe *remoteExecutor) GetName() string {
	return "Deploy Remote Controller"
}

func NewExecutorWithoutParsing(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	// Instantiate executor
	return newExecutor(namespace, controlPlane, controller), nil
}

func (exe *remoteExecutor) Execute() (err error) {
	// TODO: replace with member func
	if exe.controller.Host == "" || exe.controller.SSH.KeyFile == "" || exe.controller.SSH.User == "" {
		return util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	// Instantiate deployer
	controllerOptions := &install.ControllerOptions{
		User:            exe.controller.SSH.User,
		Host:            exe.controller.Host,
		Port:            exe.controller.SSH.Port,
		PrivKeyFilename: exe.controller.SSH.KeyFile,
		Version:         exe.controller.Package.Version,
		Repo:            exe.controller.Package.Repo,
		Token:           exe.controller.Package.Token,
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
	exe.controller.Endpoint, err = util.GetControllerEndpoint(exe.controller.Host)
	if err != nil {
		return err
	}

	return
}