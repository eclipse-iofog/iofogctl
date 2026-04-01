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

package attachexecagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name      string
	Namespace string
	Image     *string
}

type executor struct {
	name      string
	namespace string
	image     *string
}

func NewExecutor(opt Options) execute.Executor {
	return &executor{
		name:      opt.Name,
		namespace: opt.Namespace,
		image:     opt.Image,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart("Attaching Exec Session to Agent")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		msg := "%s\nFailed to get Agent by name: %s"
		return fmt.Errorf(msg, err.Error())
	}

	// Attach Exec Session to Microservice
	req := client.AttachExecToAgentRequest{
		UUID:  agent.UUID,
		Image: exe.image,
	}
	err = clt.AttachExecToAgent(&req)
	if err != nil {
		msg := "%s\nFailed to attach Exec Session to Agent: %s"
		return fmt.Errorf(msg, err.Error())
	}

	return nil
}
