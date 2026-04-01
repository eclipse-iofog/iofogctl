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

package deleterole

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	return &Executor{
		namespace: namespace,
		name:      name,
	}, nil
}

func (exe *Executor) GetName() string {
	return exe.name
}

func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Role")
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	return clt.DeleteRole(exe.name)
}
