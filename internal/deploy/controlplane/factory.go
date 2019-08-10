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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

func Deploy(opt Options) error {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Read the input file
	spec, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Instantiate executors
	var executors []execute.Executor
	for idx := range spec.Controllers {
		exe, err := newExecutor(ns.Name, spec.Controllers[idx])
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}

	// Execute
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}

	return nil
}

func newExecutor(namespace string, ctrl config.Controller) (execute.Executor, error) {
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
		return newLocalExecutor(namespace, ctrl, cli), nil
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
