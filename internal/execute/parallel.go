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

package execute

import (
	"fmt"
	"sync"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type jobResult struct {
	err error
	exe Executor
}

func RunExecutors(executors []Executor, execType string) error {
	if errs, failedExes := ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError(fmt.Sprintf("Failed to %s\n", execType))
	}
	return nil
}

func ForParallel(exes []Executor) (errs []error, failedExes []Executor) {
	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup
	// Deploy controllers
	errChan := make(chan jobResult, len(exes))
	for idx := range exes {
		wg.Add(1)
		go func(exe Executor) {
			defer wg.Done()
			err := exe.Execute()
			errChan <- jobResult{
				err: err,
				exe: exe,
			}
		}(exes[idx])
	}
	wg.Wait()
	close(errChan)

	// Output any errors
	for result := range errChan {
		if result.err != nil {
			errs = append(errs, result.err)
			failedExes = append(failedExes, result.exe)
		}
	}

	return
}
