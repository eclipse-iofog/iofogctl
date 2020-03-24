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
	ctrlClient   *client.Client
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
	name         string
}

func (exe kubernetesControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))
	if err := exe.executeInstall(); err != nil {
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
	return config.Flush()
}

func (exe kubernetesControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(namespace, name string, controlPlane *rsc.KubernetesControlPlane) execute.Executor {
	return kubernetesControlPlaneExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
		name:         name,
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

	return newControlPlaneExecutor(opt.Namespace, opt.Name, controlPlane), nil
}

func (exe *kubernetesControlPlaneExecutor) executeInstall() (err error) {
	// Get Kubernetes deployer
	installer, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetKubeletImage(exe.controlPlane.Images.Kubelet)
	installer.SetOperatorImage(exe.controlPlane.Images.Operator)
	installer.SetPortManagerImage(exe.controlPlane.Images.PortManager)
	installer.SetRouterImage(exe.controlPlane.Images.Router)
	installer.SetProxyImage(exe.controlPlane.Images.Proxy)
	installer.SetControllerImage(exe.controlPlane.Images.Controller)
	installer.SetControllerService(exe.controlPlane.Services.Controller.Type, exe.controlPlane.Services.Controller.IP)
	installer.SetRouterService(exe.controlPlane.Services.Router.Type, exe.controlPlane.Services.Router.IP)
	installer.SetRouterService(exe.controlPlane.Services.Proxy.Type, exe.controlPlane.Services.Proxy.IP)

	replicas := int32(1)
	if exe.controlPlane.Replicas.Controller != 0 {
		replicas = exe.controlPlane.Replicas.Controller
	}
	// Create controller on cluster
	if err = installer.CreateController(install.IofogUser(exe.controlPlane.IofogUser), replicas, install.Database(exe.controlPlane.Database)); err != nil {
		return
	}

	for idx := int32(0); idx < exe.controlPlane.Replicas.Controller; idx++ {
		if err := exe.controlPlane.AddController(&rsc.KubernetesController{
			PodName: fmt.Sprintf("kubernetes-%d", idx), // TODO: use actual pod name
			Created: util.NowUTC(),
		}); err != nil {
			return err
		}
	}
	// Update controller (its a pointer, this is returned to caller)
	if exe.controlPlane.Endpoint, err = installer.GetControllerEndpoint(); err != nil {
		return
	}

	return
}
