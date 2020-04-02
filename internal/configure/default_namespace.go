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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
)

type defaultNamespaceExecutor struct {
	name string
}

func newDefaultNamespaceExecutor(opt Options) *defaultNamespaceExecutor {
	return &defaultNamespaceExecutor{
		name: opt.Name,
	}
}

func (exe *defaultNamespaceExecutor) GetName() string {
	return exe.name
}

func (exe *defaultNamespaceExecutor) Execute() error {
	if err := config.SetDefaultNamespace(exe.name); err != nil {
		return err
	}

	return config.Flush()
}
