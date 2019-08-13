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

func NewExecutor(namespace string, ctrl config.Controller) (execute.Executor, error) {
	// Get the namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(ctrl.Host) {
		// Check the namespace does not contain a Controller yet
		nbControllers := len(ns.ControlPlane.Controllers)
		if nbControllers > 0 {
			return nil, util.NewInputError("This namespace already contains a Controller. Please remove it before deploying a new one.")
		}
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, ctrl, cli)
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
		return newKubernetesExecutor(namespace, ctrl), nil
	}

	// Default executor
	if ctrl.Host == "" || ctrl.KeyFile == "" || ctrl.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(namespace, ctrl), nil
}
