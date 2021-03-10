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

package deployremotecontroller

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
	controller, err := rsc.UnmarshallRemoteController(opt.Yaml)
	if err != nil {
		return
	}

	if len(opt.Name) > 0 {
		controller.Name = opt.Name
	}

	// Validate
	if err = Validate(&controller); err != nil {
		return
	}

	// Get the Control Plane
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return
	}
	controlPlane, ok := baseControlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		err = util.NewError("Could not convert Control Plane to Remote Control Plane")
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, controlPlane, &controller)
}

func newExecutor(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) *remoteExecutor {
	executor := &remoteExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
		controller:   controller,
	}

	// Set default values
	executor.setDefaultValues()
	return executor
}

func (exe *remoteExecutor) GetName() string {
	return "Deploy Remote Controller"
}

func NewExecutorWithoutParsing(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	if err := controller.Sanitize(); err != nil {
		return nil, err
	}

	if err := util.IsLowerAlphanumeric("Controller", controller.GetName()); err != nil {
		return nil, err
	}

	// Instantiate executor
	return newExecutor(namespace, controlPlane, controller), nil
}

func (exe *remoteExecutor) Execute() (err error) {
	if err = exe.controller.ValidateSSH(); err != nil {
		return
	}

	// Instantiate deployer
	controllerOptions := &install.ControllerOptions{
		User:                exe.controller.SSH.User,
		Host:                exe.controller.Host,
		Port:                exe.controller.SSH.Port,
		PrivKeyFilename:     exe.controller.SSH.KeyFile,
		PidBaseDir:          exe.controller.PidBaseDir,
		EcnViewerPort:       exe.controller.EcnViewerPort,
		Version:             exe.controlPlane.Package.Version,
		Repo:                exe.controlPlane.Package.Repo,
		Token:               exe.controlPlane.Package.Token,
		SystemMicroservices: exe.controlPlane.SystemMicroservices,
	}
	deployer, err := install.NewController(controllerOptions)
	if err != nil {
		return err
	}

	// Set database configuration
	if exe.controlPlane.Database.Host != "" {
		db := exe.controlPlane.Database
		deployer.SetControllerExternalDatabase(db.Host, db.User, db.Password, db.Provider, db.DatabaseName, db.Port)
	}

	// Deploy Controller
	if err = deployer.Install(); err != nil {
		return
	}
	// Update controller
	exe.controller.Endpoint, err = util.GetControllerEndpoint(exe.controller.Host)
	if err != nil {
		return err
	}
	return exe.controlPlane.UpdateController(exe.controller)
}

func (exe *remoteExecutor) setDefaultValues() {
	if exe.controlPlane.SystemMicroservices.Proxy.X86 == "" {
		exe.controlPlane.SystemMicroservices.Proxy.X86 = util.GetProxyImage()
	}
	if exe.controlPlane.SystemMicroservices.Proxy.ARM == "" {
		exe.controlPlane.SystemMicroservices.Proxy.ARM = util.GetProxyARMImage()
	}
	if exe.controlPlane.SystemMicroservices.Router.X86 == "" {
		exe.controlPlane.SystemMicroservices.Router.X86 = util.GetRouterImage()
	}
	if exe.controlPlane.SystemMicroservices.Router.ARM == "" {
		exe.controlPlane.SystemMicroservices.Router.ARM = util.GetRouterARMImage()
	}
}

func Validate(ctrl rsc.Controller) error {
	if err := util.IsLowerAlphanumeric("Controller", ctrl.GetName()); err != nil {
		return err
	}
	return nil
}
