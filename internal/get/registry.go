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
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
)

type registryExecutor struct {
	namespace string
}

func newRegistryExecutor(namespace string) *registryExecutor {
	a := &registryExecutor{}
	a.namespace = namespace
	return a
}

func (exe *registryExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateRegistryOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *registryExecutor) GetName() string {
	return ""
}

func generateRegistryOutput(namespace string) error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	registryList, err := clt.ListRegistries()
	if err != nil {
		return err
	}

	return tabulateRegistries(registryList.Registries)
}

func tabulateRegistries(catalogItems []client.RegistryInfo) error {
	// Generate table and headers
	table := make([][]string, len(catalogItems)+1)
	headers := []string{
		"ID",
		"URL",
		"USERNAME",
		"PRIVATE",
		"SECURE",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, item := range catalogItems {
		row := []string{
			strconv.Itoa(item.ID),
			item.URL,
			item.Username,
			strconv.FormatBool(!item.IsPublic),
			strconv.FormatBool(item.IsSecure),
		}
		table[idx+1] = append(table[idx+1], row...)
		idx++
	}

	// Print table
	return print(table)
}
