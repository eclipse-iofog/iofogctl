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
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type namespaceExecutor struct {
}

func newNamespaceExecutor() *namespaceExecutor {
	n := &namespaceExecutor{}
	return n
}

func (exe *namespaceExecutor) Execute() error {
	namespaces := config.GetNamespaces()

	// Generate table and headers
	table := make([][]string, len(namespaces)+1)
	headers := []string{"NAMESPACE", "AGE"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ns := range namespaces {

		age, err := util.Elapsed(ns.Created, util.Now())
		if err != nil {
			return err
		}
		row := []string{
			ns.Name,
			age,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print the table
	err := print(table)
	return err
}
