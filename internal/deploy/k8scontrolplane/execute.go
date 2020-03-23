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

package deployk8scontrolplane

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deploycontroller "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type kubernetesControlPlaneExecutor struct {
	ctrlClient          *client.Client
	controllerExecutors []execute.Executor
	controlPlane        *rsc.KubernetesControlPlane
	namespace           string
	name                string
}

func (exe kubernetesControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))
	if err := runExecutors(exe.controllerExecutors); err != nil {
		return err
	}

	// Make sure Controller API is ready
	endpoint, err := exe.controlPlane.GetEndpoint()
	if err != nil {
		return
	}
	if err = install.WaitForControllerAPI(endpoint); err != nil {
		return err
	}
	// Create new user
	exe.ctrlClient = client.New(client.Options{Endpoint: endpoint})
	if err = exe.ctrlClient.CreateUser(client.User(exe.controlPlane.IofogUser)); err != nil {
		// If not error about account existing, fail
		if !strings.Contains(err.Error(), "already an account associated") {
			return err
		}
		// Try to log in
		user := exe.controlPlane.GetUser()
		if err = exe.ctrlClient.Login(client.LoginRequest{
			Email:    user.Email,
			Password: user.Password,
		}); err != nil {
			return err
		}
	}
	// Update config
	if err = config.UpdateControlPlane(exe.namespace, exe.controlPlane); err != nil {
		return err
	}
	if err = config.Flush(); err != nil {
		return err
	}

	return nil
}

func (exe kubernetesControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(executors []execute.Executor, namespace, name string, controlPlane *rsc.KubernetesControlPlane) execute.Executor {
	return kubernetesControlPlaneExecutor{
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

	// Create exe Controllers
	var controllerExecutors []execute.Executor
	exe, err = deploycontroller.NewExecutorWithoutParsing(opt.Namespace, nil) // TODO: Replace this function with unique to k8s function
	if err != nil {
		return nil, err
	}
	controllerExecutors = append(controllerExecutors, exe)

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
