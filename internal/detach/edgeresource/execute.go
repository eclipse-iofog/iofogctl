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

package detachedgeresource

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type executor struct {
	nameVersion string
	namespace   string
	agent       string
}

func NewExecutor(namespace, nameVersion, agent string) execute.Executor {
	return executor{nameVersion: nameVersion, namespace: namespace, agent: agent}
}

func (exe executor) GetName() string {
	return exe.nameVersion
}

func (exe executor) Execute() error {
	util.SpinStart("Detaching Edge Resource")

	// Decode name version
	name, version, err := iutil.DecodeNameVersion(exe.nameVersion)
	if err != nil {
		return err
	}

	// Init client
	clt, err := iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Agent UUID
	agentInfo, err := clt.GetAgentByName(exe.agent, false)
	if err != nil {
		return err
	}
	// Detach from agent
	req := client.LinkEdgeResourceRequest{
		AgentUUID:           agentInfo.UUID,
		EdgeResourceName:    name,
		EdgeResourceVersion: version,
	}
	if err := clt.UnlinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
