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
	apps "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/v2/internal"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	var agent config.Agent
	if exe.useDetached {
		agent, err = config.GetDetachedAgent(exe.name)
	} else {
		agent, err = config.GetAgent(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: internal.LatestAPIVersion,
		Kind:       apps.AgentKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: agent,
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
