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

type configmapExecutor struct {
	namespace string
}

func newConfigmapExecutor(namespace string) *configmapExecutor {
	a := &configmapExecutor{}
	a.namespace = namespace
	return a
}

func (exe *configmapExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateConfigmapsOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *configmapExecutor) GetName() string {
	return ""
}

func generateConfigmapsOutput(namespace string) error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	configmapList, err := clt.ListConfigMaps()
	if err != nil {
		return err
	}

	return tabulateConfigmaps(configmapList.ConfigMaps)
}

func tabulateConfigmaps(configMaps []client.ConfigMapInfo) error {
	// Generate table and headers
	table := make([][]string, len(configMaps)+1)
	headers := []string{
		"CONFIGMAP",
		"ID",
		"IMMUTABLE",
	}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, configMap := range configMaps {

		row := []string{
			configMap.Name,
			strconv.Itoa(configMap.ID),
			strconv.FormatBool(configMap.Immutable),
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	return print(table)
}
