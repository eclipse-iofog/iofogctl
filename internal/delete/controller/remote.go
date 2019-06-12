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
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type remoteExecutor struct {
	namespace string
	name      string
}

func newRemoteExecutor(namespace, name string) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *remoteExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := config.DeleteController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
