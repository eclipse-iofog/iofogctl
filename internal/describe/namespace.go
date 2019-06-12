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

package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type namespaceExecutor struct {
	name string
}

func newNamespaceExecutor(name string) *namespaceExecutor {
	n := &namespaceExecutor{}
	n.name = name
	return n
}

func (exe *namespaceExecutor) Execute() error {
	namespace, err := config.GetNamespace(exe.name)
	if err != nil {
		return err
	}
	if err = print(namespace); err != nil {
		return err
	}
	return nil
}
