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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace string, ctrl *config.Controller, controlPlane config.ControlPlane) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric(ctrl.Name); err != nil {
		return nil, err
	}

	if controlPlane.IofogUser.Email == "" || controlPlane.IofogUser.Password == "" {
		return nil, util.NewError("Cannot deploy Controller because ioFog user is not specified")
	}
	// Local executor
	if util.IsLocalHost(ctrl.Host) {
		// Check the namespace does not contain a Controller yet
		nbControllers := len(controlPlane.Controllers)
		if nbControllers != 1 {
			return nil, util.NewInputError("Cannot deploy more than a single Controller locally")
		}
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, ctrl, controlPlane, cli)
	}

	// Kubernetes executor
	if ctrl.KubeConfig != "" {
		// TODO: re-enable specifying images
		// If image file specified, read it
		//if ctrl.ImagesFile != "" {
		//	ctrl.Images = make(map[string]string)
		//	err := util.UnmarshalYAML(opt.ImagesFile, opt.Images)
		//	if err != nil {
		//		return nil, err
		//	}
		//}
		return newKubernetesExecutor(namespace, ctrl, controlPlane), nil
	}

	// Default executor
	if ctrl.Host == "" || ctrl.KeyFile == "" || ctrl.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(namespace, ctrl, controlPlane), nil
}
