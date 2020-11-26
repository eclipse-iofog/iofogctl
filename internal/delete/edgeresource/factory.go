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

package deleteedgeresource

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
)

type executor struct {
	namespace   string
	nameVersion string
}

func (exe executor) GetName() string {
	return "deleting Edge Resource " + exe.nameVersion
}

func (exe executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Decode nameVersion
	name, version, err := iutil.DecodeNameVersion(exe.nameVersion)
	if err != nil {
		return err
	}

	// Connect to Controller
	clt, err := iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	// Check capability
	if err := iutil.IsEdgeResourceCapable(exe.namespace); err != nil {
		return err
	}

	if err = clt.DeleteEdgeResource(name, version); err != nil {
		return
	}
	return
}

func NewExecutor(namespace, nameVersion string) (exe execute.Executor) {
	return executor{
		namespace:   namespace,
		nameVersion: nameVersion,
	}
}
