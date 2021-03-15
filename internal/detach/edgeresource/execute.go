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
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type executor struct {
	name      string
	version   string
	namespace string
	agent     string
}

func NewExecutor(namespace, name, version, agent string) execute.Executor {
	return executor{name: name,
		version:   version,
		namespace: namespace,
		agent:     agent}
}

func (exe executor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.name, exe.version)
}

func (exe executor) Execute() error {
	util.SpinStart("Detaching Edge Resource")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
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
		EdgeResourceName:    exe.name,
		EdgeResourceVersion: exe.version,
	}
	if err := clt.UnlinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
