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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	controller config.Controller
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
			controller.KeyFile = exe.controller.KeyFile
			controller.Port = exe.controller.Port
			controller.User = exe.controller.User
			config.UpdateController(exe.namespace, controller)
			return nil
		}
	}

	util.PrintNotify(fmt.Sprintf("ECN does not contain controller %s\n", exe.controller.Name))
	return nil
}

func NewExecutor(name, namespace string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	controller, err := deploycontroller.UnmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}
	controller.Name = name

	return executor{namespace: namespace, controller: controller}, nil
}
