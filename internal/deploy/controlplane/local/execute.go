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

package deploylocalcontrolplane

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/agentconfig"
	"github.com/eclipse-iofog/iofogctl/v2/internal/deploy/controller/local"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}
type localControlPlaneExecutor struct {
	ctrlClient          *client.Client
	controllerExecutors []execute.Executor
	controlPlane        rsc.ControlPlane
	namespace           string
	name                string
}

// TODO: remove duplication
func deploySystemAgent(namespace string, ctrl rsc.Controller) (err error) {
	host := "localhost"
	// Deploy system agent to host internal router
	install.Verbose("Deploying system agent")
	agent := rsc.Agent{
		Name: iofog.VanillaRouterAgentName,
		Host: host,
	}
	// Configure agent to be system agent with default router
	RouterConfig := client.RouterConfig{
		RouterMode:      internal.MakeStrPtr("interior"),
		MessagingPort:   internal.MakeIntPtr(5672),
		EdgeRouterPort:  internal.MakeIntPtr(56721),
		InterRouterPort: internal.MakeIntPtr(56722),
	}
	deployAgentConfig := rsc.AgentConfiguration{
		Name: iofog.VanillaRouterAgentName,
		AgentConfiguration: client.AgentConfiguration{
			IsSystem:     internal.MakeBoolPtr(true),
			Host:         &host,
			RouterConfig: RouterConfig,
		},
	}

	// Get Agentconfig executor
	deployAgentConfigExecutor := deployagentconfig.NewRemoteExecutor(iofog.VanillaRouterAgentName, deployAgentConfig, namespace)
	// If there already is a system fog, ignore error
	if err = deployAgentConfigExecutor.Execute(); err != nil {
		if strings.Contains(err.Error(), "There already is a system fog") {
			util.PrintNotify(fmt.Sprintf("Using existing default router"))
			return nil
		}
		return err
	}
	agent.UUID = deployAgentConfigExecutor.GetAgentUUID()
	return nil
}

// TODO: remove duplication
func (exe localControlPlaneExecutor) postDeploy() (err error) {
	// Look for a Vanilla controller
	controllers := exe.controlPlane.GetControllers()
	for _, ctrl := range controllers {
		// If Vanilla controller
		if err = deploySystemAgent(exe.namespace, ctrl); err != nil {
			return err
		}
	}
	return nil
}

func (exe localControlPlaneExecutor) Execute() (err error) {
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
	if err = exe.ctrlClient.CreateUser(client.User(exe.controlPlane.GetUser())); err != nil {
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
	config.UpdateControlPlane(exe.namespace, exe.controlPlane)
	if err = config.Flush(); err != nil {
		return err
	}
	// Post deploy steps
	return exe.postDeploy()
}

func (exe localControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(executors []execute.Executor, namespace, name string, controlPlane rsc.ControlPlane) execute.Executor {
	return localControlPlaneExecutor{
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
	for _, baseController := range controlPlane.GetControllers() {
		controller, ok := baseController.(*rsc.LocalController)
		if !ok {
			return nil, util.NewError("Could not convert Controller to Local Controller")
		}
		exe, err := deploylocalcontroller.NewExecutorWithoutParsing(opt.Namespace, &controlPlane, controller)
		if err != nil {
			return nil, err
		}
		controllerExecutors = append(controllerExecutors, exe)
	}

	return newControlPlaneExecutor(controllerExecutors, opt.Namespace, opt.Name, &controlPlane), nil
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
