/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
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
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type roleBindingExecutor struct {
	namespace string
}

func newRoleBindingExecutor(namespace string) *roleBindingExecutor {
	a := &roleBindingExecutor{}
	a.namespace = namespace
	return a
}

func (exe *roleBindingExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateRoleBindingsOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *roleBindingExecutor) GetName() string {
	return ""
}

func generateRoleBindingsOutput(namespace string) error {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	list, err := clt.ListRoleBindings()
	if err != nil {
		return err
	}

	return tabulateRoleBindings(list.Bindings)
}

func tabulateRoleBindings(bindings []client.RoleBindingInfo) error {
	table := make([][]string, len(bindings)+1)
	headers := []string{
		"NAME",
		"KIND",
		"ROLE",
		"SUBJECTS",
	}
	table[0] = append(table[0], headers...)

	for idx, b := range bindings {
		roleName := b.RoleRef.Name
		subjectsCount := strconv.Itoa(len(b.Subjects))
		row := []string{
			b.Name,
			b.Kind,
			roleName,
			subjectsCount,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return print(table)
}
