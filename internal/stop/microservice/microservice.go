/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package microservice

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type Options struct {
	Namespace string
	Name      string
}

type executor struct {
	namespace string
	name      string
}

func NewExecutor(opt Options) (exe execute.Executor) {
	return &executor{
		name:      opt.Name,
		namespace: opt.Namespace,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() (err error) {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	appName, msvcName, err := clientutil.ParseFQName(exe.name, "Microservice")
	if err != nil {
		return err
	}

	response, err := clt.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		return
	}

	uuid := response.UUID

	err = clt.StopMicroservice(uuid)

	return
}
