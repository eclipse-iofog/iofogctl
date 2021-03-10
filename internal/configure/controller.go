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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type remoteConfig struct {
	keyFile string
	user    string
	port    int
}

type controllerExecutor struct {
	namespace        string
	kubernetesConfig kubernetesConfig
	name             string
	remoteConfig     remoteConfig
}

func newControllerExecutor(opt *Options) *controllerExecutor {
	return &controllerExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		remoteConfig: remoteConfig{
			keyFile: opt.KeyFile,
			user:    opt.User,
			port:    opt.Port,
		},
	}
}

func (exe *controllerExecutor) GetName() string {
	return exe.name
}

func (exe *controllerExecutor) Execute() error {
	// Get config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	switch controlPlane := baseControlPlane.(type) {
	case *rsc.RemoteControlPlane:
		if err := exe.remoteConfigure(controlPlane); err != nil {
			return err
		}

	case *rsc.KubernetesControlPlane:
		return util.NewInputError("Cannot configure Kubernetes Control Plane as if it is a Remote Control Plane")

	case *rsc.LocalControlPlane:
		return util.NewInputError("Cannot configure a Local ControlPlane")
	}

	ns.SetControlPlane(baseControlPlane)
	return config.Flush()
}

func (exe *controllerExecutor) remoteConfigure(controlPlane *rsc.RemoteControlPlane) (err error) {
	// Error if kubernetesConfig is passed
	if (kubernetesConfig{}) != exe.kubernetesConfig {
		return util.NewInputError("Cannot edit Kubernetes config of a Remote ControlPlane")
	}

	baseController, err := controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}
	controller, ok := baseController.(*rsc.RemoteController)
	if !ok {
		return util.NewInternalError("Failed to convert Controller into Remote Controller")
	}
	if exe.remoteConfig.keyFile != "" {
		controller.SSH.KeyFile, err = util.FormatPath(exe.remoteConfig.keyFile)
		if err != nil {
			return err
		}
	}
	if exe.remoteConfig.user != "" {
		controller.SSH.User = exe.remoteConfig.user
	}
	if exe.remoteConfig.port != 0 {
		controller.SSH.Port = exe.remoteConfig.port
	}

	// Add port if not specified or existing
	if controller.Host != "" && controller.SSH.Port == 0 {
		controller.SSH.Port = 22
	}

	if err := controlPlane.UpdateController(controller); err != nil {
		return err
	}
	return nil
}
