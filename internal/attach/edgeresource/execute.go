/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package attachedgeresource

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type Options struct {
	Name      string
	Version   string
	Agent     string
	Namespace string
}

type executor struct {
	Options
}

func NewExecutor(opt Options) execute.Executor {
	return executor{opt}
}

func (exe executor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.Name, exe.Version)
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Edge Resource")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.Namespace)
	if err != nil {
		return err
	}

	// Get Agent UUID
	agentInfo, err := clt.GetAgentByName(exe.Agent, false)
	if err != nil {
		return err
	}
	// Attach to agent
	req := client.LinkEdgeResourceRequest{
		AgentUUID:           agentInfo.UUID,
		EdgeResourceName:    exe.Name,
		EdgeResourceVersion: exe.Version,
	}
	if err := clt.LinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
