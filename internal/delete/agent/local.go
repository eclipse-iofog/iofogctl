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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	namespace        string
	client           *iofog.LocalContainer
	localAgentConfig *iofog.LocalAgentConfig
}

func newLocalExecutor(namespace, name string, client *iofog.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:        namespace,
		client:           client,
		localAgentConfig: iofog.NewLocalAgentConfig(name),
	}
	return exe
}

func (exe *localExecutor) Execute() error {
	// Clean all agent containers
	if errClean := exe.client.CleanContainer(exe.localAgentConfig.ContainerName); errClean != nil {
		fmt.Printf("Could not clean Agent container: %v", errClean)
	}

	// Update configuration
	err := config.DeleteAgent(exe.namespace, exe.localAgentConfig.Name)
	if err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.localAgentConfig.Name)

	return config.Flush()
}
