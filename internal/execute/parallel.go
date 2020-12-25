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
	"errors"
	"fmt"
	"sync"
)

type jobResult struct {
	err error
	exe Executor
}

func CoalesceErrors(errs []error) error {
	msg := ""
	for idx := range errs {
		if msg == "" {
			msg = errs[idx].Error()
		} else {
			msg = fmt.Sprintf("%s\n%s", msg, errs[idx].Error())
		}
	}
	return errors.New(msg)
}

func RunExecutors(executors []Executor, execType string) []error {
	if errs, _ := ForParallel(executors); len(errs) > 0 {
		return errs
	}
	return []error{}
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
