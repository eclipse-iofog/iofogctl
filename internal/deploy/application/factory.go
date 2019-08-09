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

package deployapplication

import (
	"fmt"
	"sync"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor interface {
	execute() error
}

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

	// Check Controller exists
	nbControllers := len(ns.ControlPlane.Controllers)
	if nbControllers != 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", nbControllers)
		return util.NewInputError(errMessage)
	}

	applications, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup
	errChan := make(chan jobResult, len(applications))

	// Deploy applications
	for _, application := range applications {
		exe, err := newExecutor(opt.Namespace, &application)
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
		}(application.Name)
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

func newExecutor(namespace string, opt *config.Application) (executor, error) {
	return newRemoteExecutor(namespace, opt), nil
}
