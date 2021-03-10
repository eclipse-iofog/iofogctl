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

package describe

import (
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type agentExecutor struct {
	namespace   string
	name        string
	filename    string
	useDetached bool
}

func newAgentExecutor(namespace, name, filename string, useDetached bool) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	a.useDetached = useDetached
	return a
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() (err error) {
	var agent rsc.Agent
	if exe.useDetached {
		agent, err = config.GetDetachedAgent(exe.name)
		if err != nil {
			return err
		}
	} else {
		ns, err := config.GetNamespace(exe.namespace)
		if err != nil {
			return err
		}
		// Update local cache based on Controller
		if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
			return err
		}
		agent, err = ns.GetAgent(exe.name)
		if err != nil {
			return err
		}
	}

	var tags *[]string
	var agentConfig rsc.AgentConfiguration
	// Detached Agents don't have a UUID
	if agent.GetUUID() != "" {
		// Get Agent configuration
		agentConfig, tags, err = clientutil.GetAgentConfig(exe.name, exe.namespace)
		if err != nil {
			return err
		}
		agent.SetConfig(&agentConfig)
	}

	var kind config.Kind
	switch agent.(type) {
	case *rsc.LocalAgent:
		kind = config.LocalAgentKind
	case *rsc.RemoteAgent:
		kind = config.RemoteAgentKind
	}
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       kind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
			Tags:      tags,
		},
		Spec: agent,
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
