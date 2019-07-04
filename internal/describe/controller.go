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

package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
}

func newControllerExecutor(namespace, name string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	c.name = name
	return c
}

func (exe *controllerExecutor) Execute() error {
	controller, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	println("namespace: " + exe.namespace)
	controller.IofogUser.Password = "*****"
	if err = util.Print(controller); err != nil {
		return err
	}
	return nil
}
