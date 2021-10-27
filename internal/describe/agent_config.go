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
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type agentConfigExecutor struct {
	namespace string
	name      string
	filename  string
}

func newAgentConfigExecutor(namespace, name, filename string) *agentConfigExecutor {
	a := &agentConfigExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *agentConfigExecutor) GetName() string {
	return exe.name
}

func (exe *agentConfigExecutor) Execute() error {
	agentConfig, tags, err := clientutil.GetAgentConfig(exe.name, exe.namespace)
	if err != nil {
		return err
	}
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.AgentConfigKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
			Tags:      tags,
		},
		Spec: agentConfig,
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
