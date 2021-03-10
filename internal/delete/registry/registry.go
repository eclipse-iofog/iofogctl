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

package deleteregistry

import (
	"strconv"

	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type Executor struct {
	namespace string
	id        int
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	id, err := strconv.Atoi(name)
	if err != nil {
		return nil, err
	}
	exe := &Executor{
		namespace: namespace,
		id:        id,
	}

	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return strconv.Itoa(exe.id)
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Registry")
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	return clt.DeleteRegistry(exe.id)
}
