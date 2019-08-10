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

package execute

import (
	"sync"
)

type jobResult struct {
	err error
	exe Executor
}

func ForParallel(exes []Executor) (errs []error, failedExes []Executor) {
	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup
	// Deploy controllers
	errChan := make(chan jobResult, len(exes))
	for idx := range exes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := exes[idx].Execute()
			errChan <- jobResult{
				err: err,
				exe: exes[idx],
			}
		}()
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
