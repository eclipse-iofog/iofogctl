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

package deploymicroservice

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	deploytypes "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy/microservice"
)

type remoteExecutor struct {
	microservice deploytypes.Microservice
	controller   deploytypes.IofogController
}

func (exe remoteExecutor) GetName() string {
	return exe.microservice.Name
}

func (exe remoteExecutor) Execute() error {
	return deploy.Execute(exe.controller, exe.microservice)
}

func newExecutor(controller deploytypes.IofogController, opt deploytypes.Microservice) (execute.Executor, error) {
	return remoteExecutor{
		controller:   controller,
		microservice: opt,
	}, nil
}
