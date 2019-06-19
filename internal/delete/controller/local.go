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

package deletecontroller

import (
	"fmt"

	pb "github.com/schollz/progressbar"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	namespace             string
	name                  string
	client                *iofog.LocalContainer
	localControllerConfig *iofog.LocalControllerConfig
	pb                    *pb.ProgressBar
}

func newLocalExecutor(namespace, name string, client *iofog.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:             namespace,
		name:                  name,
		client:                client,
		localControllerConfig: iofog.NewLocalControllerConfig(name, make(map[string]string)),
		pb:                    pb.New(100),
	}
	return exe
}

func (exe *localExecutor) Execute() error {
	exe.pb.Add(1)
	defer exe.pb.Clear()
	// Clean controller and connector containers
	exe.pb.Add(10)
	for _, containerConfig := range exe.localControllerConfig.ContainerMap {
		if errClean := exe.client.CleanContainer(containerConfig.ContainerName); errClean != nil {
			fmt.Printf("Could not clean Controller container: %v", errClean)
		}
		exe.pb.Add(35)
	}

	// Update configuration
	err := config.DeleteController(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	exe.pb.Add(19)

	fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
