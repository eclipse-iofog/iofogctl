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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
)

type volumeExecutor struct {
	namespace string
}

func newVolumeExecutor(namespace string) *volumeExecutor {
	c := &volumeExecutor{}
	c.namespace = namespace
	return c
}

func (exe *volumeExecutor) GetName() string {
	return ""
}

func (exe *volumeExecutor) Execute() error {
	printNamespace(exe.namespace)
	return generateVolumeOutput(exe.namespace)
}

func generateVolumeOutput(namespace string) error {
	// Get volume config details
	volumes, err := config.GetVolumes(namespace)
	if err != nil {
		return err
	}

	// Generate table and headers
	table := make([][]string, len(volumes)+1)
	headers := []string{"VOLUME", "SOURCE", "DESTINATION", "PERMISSIONS", "AGENTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, volume := range volumes {
		// Create list of Agents
		agentList := ""
		for idx, agent := range volume.Agents {
			separator := ", "
			if idx == 0 {
				separator = ""
			}
			agentList = agentList + separator + agent
		}
		// Store values
		row := []string{
			volume.Name,
			volume.Source,
			volume.Destination,
			volume.Permissions,
			agentList,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
