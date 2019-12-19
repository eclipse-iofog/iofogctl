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

package deploycontrolplane

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type controlPlaneExecutor struct {
	controllerExecutors []execute.Executor
	controlPlane        config.ControlPlane
	namespace           string
	name                string
}

func (exe controlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))
	if err := runExecutors(exe.controllerExecutors); err != nil {
		return err
	}

	// Make sure Controller API is ready
	endpoint, err := exe.controlPlane.GetControllerEndpoint()
	if err != nil {
		return
	}
	if err = install.WaitForControllerAPI(endpoint); err != nil {
		return err
	}
	// Create new user
	ctrlClient := client.New(endpoint)
	if err = ctrlClient.CreateUser(client.User(exe.controlPlane.IofogUser)); err != nil {
		// If not error about account existing, fail
		if !strings.Contains(err.Error(), "already an account associated") {
			return err
		}
		// Try to log in
		if err = ctrlClient.Login(client.LoginRequest{
			Email:    exe.controlPlane.IofogUser.Email,
			Password: exe.controlPlane.IofogUser.Password,
		}); err != nil {
			return err
		}
	}
	// Update config
	if err = config.UpdateControlPlane(exe.namespace, exe.controlPlane); err != nil {
		return err
	}
	return config.Flush()
}

func (exe controlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(executors []execute.Executor, namespace, name string, controlPlane config.ControlPlane) execute.Executor {
	return controlPlaneExecutor{
		controllerExecutors: executors,
		namespace:           namespace,
		controlPlane:        controlPlane,
		name:                name,
	}
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	_, err = config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	// Read the input file
	controlPlane, err := UnmarshallYAML(opt.Yaml)
	if err != nil {
		return
	}

	// Instantiate executors
	var controllerExecutors []execute.Executor

	// Create exe Controllers
	for idx := range controlPlane.Controllers {
		exe, err := deploycontroller.NewExecutorWithoutParsing(opt.Namespace, &controlPlane.Controllers[idx], controlPlane)
		if err != nil {
			return nil, err
		}
		controllerExecutors = append(controllerExecutors, exe)
	}

	return newControlPlaneExecutor(controllerExecutors, opt.Namespace, opt.Name, controlPlane), nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}
	return nil
}
