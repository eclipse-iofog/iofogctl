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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type kubernetesConfig struct {
	kubeConfig string
}

type remoteConfig struct {
	host    string
	keyFile string
	user    string
	port    int
}

type controlPlaneExecutor struct {
	namespace        string
	kubernetesConfig kubernetesConfig
	name             string
	remoteConfig     remoteConfig
}

func newControlPlaneExecutor(opt Options) *controlPlaneExecutor {
	return &controlPlaneExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		remoteConfig: remoteConfig{
			keyFile: opt.KeyFile,
			user:    opt.User,
			port:    opt.Port,
			host:    opt.Host,
		},
		kubernetesConfig: kubernetesConfig{
			kubeConfig: opt.KubeConfig,
		},
	}
}

func (exe *controlPlaneExecutor) GetName() string {
	return exe.name
}

func (exe *controlPlaneExecutor) Execute() error {
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
		if err = exe.remoteConfigure(controlPlane); err != nil {
			return err
		}

	case *rsc.KubernetesControlPlane:
		if err = exe.kubernetesConfigure(controlPlane); err != nil {
			return err
		}

	case *rsc.LocalControlPlane:
		return util.NewInputError("Cannot configure a Local Controlplane")
	}

	config.UpdateControlPlane(exe.namespace, baseControlPlane)

	return config.Flush()
}

func (exe *controlPlaneExecutor) kubernetesConfigure(controlPlane *rsc.KubernetesControlPlane) (err error) {
	// Error if remoteConfig is passed
	if (remoteConfig{}) != exe.remoteConfig {
		return util.NewInputError("Cannot edit remote config of a kubernetes controlplane")
	}

	if exe.kubernetesConfig.kubeConfig != "" {
		controlPlane.KubeConfig = exe.kubernetesConfig.kubeConfig
	}

	return nil
}

func (exe *controlPlaneExecutor) remoteConfigure(controlPlane *rsc.RemoteControlPlane) (err error) {
	// Error if kubernetesConfig is passed
	if (kubernetesConfig{}) != exe.kubernetesConfig {
		return util.NewInputError("Cannot edit kubernetes config of a remote controlplane")
	}

	controllers := controlPlane.GetControllers()

	// TODO: Find a way to allow configuration of separate controllers
	for _, baseController := range controllers {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Failed to convert controller into remoteController")
		}

		// Disallow editing host if already exists
		if controller.Host != "" && exe.remoteConfig.host != "" {
			return util.NewInputError("Cannot edit existing host address of Controller. Can only add host address where it doesn't exist after running connect command")
		}
		// Only add/overwrite values provided
		if exe.remoteConfig.host != "" {
			controller.Host = exe.remoteConfig.host
			controller.Endpoint, err = util.GetControllerEndpoint(exe.remoteConfig.host)
			if err != nil {
				return err
			}
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

		if err = controlPlane.UpdateController(baseController); err != nil {
			return err
		}
	}
	return nil
}
