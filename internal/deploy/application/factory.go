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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

func NewExecutor(namespace string, opt *config.Application) (Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}

	// Check Controller exists
	nbControllers := len(ns.Controllers)
	if nbControllers != 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", nbControllers)
		return nil, util.NewInputError(errMessage)
	}

	return newRemoteExecutor(namespace, opt), nil
}
