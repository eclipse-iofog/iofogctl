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

	pb "github.com/schollz/progressbar"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	namespace        string
	client           *iofog.LocalContainer
	localAgentConfig *iofog.LocalAgentConfig
	pb               *pb.ProgressBar
}

func newLocalExecutor(namespace, name string, client *iofog.LocalContainer) *localExecutor {
	ctrlConfig, _ := iofog.NewLocalControllerConfig("", make(map[string]string)).ContainerMap["controller"]
	exe := &localExecutor{
		namespace:        namespace,
		client:           client,
		localAgentConfig: iofog.NewLocalAgentConfig(name, "", ctrlConfig),
		pb:               pb.New(100),
	}
	return exe
}

func (exe *localExecutor) Execute() error {
	exe.pb.Add(1)
	defer exe.pb.Clear()
	// Clean all agent containers
	exe.pb.Add(10)
	if errClean := exe.client.CleanContainer(exe.localAgentConfig.ContainerName); errClean != nil {
		fmt.Printf("Could not clean Agent container: %v", errClean)
	}
	exe.pb.Add(70)

	// Update configuration
	err := config.DeleteAgent(exe.namespace, exe.localAgentConfig.Name)
	if err != nil {
		return err
	}
	exe.pb.Add(19)

	fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.localAgentConfig.Name)

	return config.Flush()
}
