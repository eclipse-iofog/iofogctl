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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
	filename  string
}

func newControllerExecutor(namespace, name, filename string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	c.name = name
	c.filename = filename
	return c
}

func (exe *controllerExecutor) Execute() error {
	controller, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	fmt.Printf("namespace: %s\n", exe.namespace)
	if exe.filename == "" {
		if err = util.Print(controller); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(controller, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
