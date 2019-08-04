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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
	client    *client.Client
	flow      *client.FlowInfo
}

func NewExecutor(namespace, name string) *Executor {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}

	return exe
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() (err error) {
	// Get Controllers from namespace
	controllers, err := config.GetControllers(exe.namespace)

	// Do we actually have any controllers?
	if err != nil {
		util.PrintError("No controller found in this namespace")
		return
	}

	// Did we have more than one controller?
	if len(controllers) != 1 {
		err = util.NewInternalError("Only support 1 controller per namespace")
		return
	}

	// Init remote resources
	if err = exe.init(&controllers[0]); err != nil {
		return
	}

	// Delete flow
	if err = exe.client.DeleteFlow(exe.flow.ID); err != nil {
		return
	}

	return nil
}

func (exe *Executor) init(controller *config.Controller) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controller.IofogUser.Email, Password: controller.IofogUser.Password}); err != nil {
		return
	}
	flow, err := exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	exe.flow = flow
	return
}
