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

package connectcontroller

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type executor struct {
	controller rsc.Controller
	namespace  string
}

func (exe executor) GetName() string {
	return exe.controller.Name
}

func (exe executor) Execute() error {
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}

	for _, controller := range controllers {
		if controller.Name == exe.controller.Name {
			// Only update ssh info
			controller.SSH.KeyFile = exe.controller.SSH.KeyFile
			controller.SSH.Port = exe.controller.SSH.Port
			controller.SSH.User = exe.controller.SSH.User
			config.UpdateController(exe.namespace, controller)
			return nil
		}
	}

	util.PrintNotify(fmt.Sprintf("ECN does not contain controller %s\n", exe.controller.Name))
	return nil
}

func NewExecutor(namespace, name string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	controller, err := unmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}
	controller.Name = name

	return executor{namespace: namespace, controller: controller}, nil
}
