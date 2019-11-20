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
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentConfigExecutor struct {
	namespace string
	name      string
	filename  string
}

func newAgentConfigExecutor(namespace, name, filename string) *agentConfigExecutor {
	a := &agentConfigExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *agentConfigExecutor) GetName() string {
	return exe.name
}

func (exe *agentConfigExecutor) Execute() error {
	// Get config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Connect to controller
	ctrl, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	getAgentResponse, err := ctrl.GetAgentByID(agent.UUID)
	if err != nil {
		// The agents might not be provisioned with Controller
		if strings.Contains(err.Error(), "NotFoundError") {
			return util.NewInputError("Cannot describe an Agent that is not provisioned with the Controller in namespace " + exe.namespace)
		}
		return err
	}

	fogType, found := config.FogTypeIntMap[getAgentResponse.FogType]
	if !found {
		fogType = "auto"
	}

	agentConfig := config.AgentConfiguration{
		Name:        getAgentResponse.Name,
		Location:    getAgentResponse.Location,
		Latitude:    getAgentResponse.Latitude,
		Longitude:   getAgentResponse.Longitude,
		Description: getAgentResponse.Description,
		FogType:     fogType,
		AgentConfiguration: client.AgentConfiguration{
			DockerURL:                 &getAgentResponse.DockerURL,
			DiskLimit:                 &getAgentResponse.DiskLimit,
			DiskDirectory:             &getAgentResponse.DiskDirectory,
			MemoryLimit:               &getAgentResponse.MemoryLimit,
			CPULimit:                  &getAgentResponse.CPULimit,
			LogLimit:                  &getAgentResponse.LogLimit,
			LogDirectory:              &getAgentResponse.LogDirectory,
			LogFileCount:              &getAgentResponse.LogFileCount,
			StatusFrequency:           &getAgentResponse.StatusFrequency,
			ChangeFrequency:           &getAgentResponse.ChangeFrequency,
			DeviceScanFrequency:       &getAgentResponse.DeviceScanFrequency,
			BluetoothEnabled:          &getAgentResponse.BluetoothEnabled,
			WatchdogEnabled:           &getAgentResponse.WatchdogEnabled,
			AbstractedHardwareEnabled: &getAgentResponse.AbstractedHardwareEnabled,
		},
	}

	header := config.Header{
		APIVersion: internal.LatestAPIVersion,
		Kind:       config.AgentConfigKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: agentConfig,
	}

	if exe.namespace == "" {
		header.Metadata.Namespace = config.GetCurrentNamespace().Name
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
