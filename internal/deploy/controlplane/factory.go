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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"sync"
)

type Options struct {
	Namespace string
	InputFile string
}

type jobResult struct {
	name string
	err  error
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

	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup
	// Deploy controllers
	errChan := make(chan jobResult, len(spec.Controllers))
	for idx := range spec.Controllers {
		var exe executor
		exe, err = newExecutor(ns.Name, spec.Controllers[idx])
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			err := exe.execute()
			errChan <- jobResult{
				err:  err,
				name: name,
			}
		}(spec.Controllers[idx].Name)
	}
	wg.Wait()
	close(errChan)

	// Output any errors
	failed := false
	for result := range errChan {
		if result.err != nil {
			failed = true
			util.PrintNotify("Failed to deploy " + result.name + ". " + result.err.Error())
		}
	}

	if failed {
		return util.NewError("Failed to deploy one or more resources")
	}

	return nil
}

type executor interface {
	execute() error
}

func newExecutor(namespace string, ctrl config.Controller) (executor, error) {
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
