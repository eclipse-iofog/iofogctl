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

package deleteapplicationtemplate

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func Execute(namespace, name string) error {
	// Get executor
	exe := NewExecutor(namespace, name)

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	// Leave this here as a note on general practice with Execute functions
	return config.Flush()
}

type Executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string) execute.Executor {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}

	return exe
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.name
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Application Template")
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if err := clt.DeleteApplicationTemplate(exe.name); err != nil {
		return err
	}

	return nil
}
