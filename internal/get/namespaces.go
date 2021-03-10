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

package get

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type namespaceExecutor struct {
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	return n
}

func (exe *namespaceExecutor) GetName() string {
	return ""
}

func (exe *namespaceExecutor) Execute() error {
	namespacesNames := config.GetNamespaces()
	namespaces := make([]*rsc.Namespace, len(namespacesNames))
	for idx, n := range namespacesNames {
		ns, err := config.GetNamespace(n)
		if err != nil {
			return err
		}
		namespaces[idx] = ns
	}

	// Generate table and headers
	table := make([][]string, len(namespaces))
	headers := []string{"NAMESPACE", "AGE"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ns := range namespaces {
		age, err := util.ElapsedUTC(ns.Created, util.NowUTC())
		if err != nil {
			age = "-"
		}
		row := []string{
			ns.Name,
			age,
		}
		if ns.Name == config.GetDefaultNamespaceName() {
			row[0] = ns.Name + "*"
			prepend := [][]string{table[0]}
			table = append(prepend, table...)
			table[1] = row
		} else {
			table[idx+1] = append(table[idx+1], row...)
		}
	}

	// Print the table
	return print(table)
}
