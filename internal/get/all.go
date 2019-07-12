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

package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type allExecutor struct {
	namespace string
}

func newAllExecutor(namespace string) *allExecutor {
	exe := &allExecutor{}
	exe.namespace = namespace
	return exe
}

func (exe *allExecutor) Execute() error {
	// Check namespace exists
	if _, err := config.GetNamespace(exe.namespace); err != nil {
		return err
	}
	printNamespace(exe.namespace)

	// Print controllers
	if err := generateControllerOutput(exe.namespace); err != nil {
		return err
	}

	// Print agents
	if err := generateAgentOutput(exe.namespace); err != nil {
		return err
	}

	return nil
}
