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

package attachagent

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Options struct {
	NameVersion string
	Agent       string
	Namespace   string
}

type executor struct {
	opt Options
}

func NewExecutor(opt Options) execute.Executor {
	return executor{opt: opt}
}

func (exe executor) GetName() string {
	return exe.opt.NameVersion
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Edge Resource")

	// Decode name version
	name, version, err := iutil.DecodeNameVersion(exe.opt.NameVersion)
	if err != nil {
		return err
	}

	// Init remote resources
	clt, err := iutil.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return err
	}

	// Get Agent UUID
	agentInfo, err := clt.GetAgentByName(exe.opt.Agent, false)
	if err != nil {
		return err
	}
	// Attach to agent
	req := client.LinkEdgeResourceRequest{
		AgentUUID:           agentInfo.UUID,
		EdgeResourceName:    name,
		EdgeResourceVersion: version,
	}
	if err := clt.LinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
