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

package deleteagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace        string
	name             string
	client           *install.LocalContainer
	localAgentConfig *install.LocalAgentConfig
}

func newLocalExecutor(namespace, name string, client *install.LocalContainer) *localExecutor {
	ctrlConfig := install.NewLocalControllerConfig(make(map[string]string), install.Credentials{})
	exe := &localExecutor{
		namespace:        namespace,
		name:             name,
		client:           client,
		localAgentConfig: install.NewLocalAgentConfig(name, "", ctrlConfig, install.Credentials{}),
	}
	return exe
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) Execute() error {
	// Clean all agent containers
	if errClean := exe.client.CleanContainer(exe.localAgentConfig.ContainerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Agent container: %v", errClean))
	}

	return nil
}
