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

package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"

	"github.com/eclipse-iofog/iofogctl/internal"
	"gopkg.in/yaml.v2"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	name        string
	agentConfig config.AgentConfiguration
	namespace   string
}

func (exe remoteExecutor) GetName() string {
	return exe.name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying agent %s configuration", exe.GetName()))

	// Check controller is reachable
	clt, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		return err
	}

	fogType, found := config.FogTypeStringMap[exe.agentConfig.FogType]
	if !found {
		fogType = 0
	}

	updateAgentConfigRequest := client.AgentUpdateRequest{
		UUID:               agent.UUID,
		Location:           exe.agentConfig.Location,
		Latitude:           exe.agentConfig.Latitude,
		Longitude:          exe.agentConfig.Longitude,
		Description:        exe.agentConfig.Description,
		FogType:            fogType,
		Name:               exe.agentConfig.Name,
		AgentConfiguration: exe.agentConfig.AgentConfiguration,
	}

	if _, err = clt.UpdateAgent(&updateAgentConfigRequest); err != nil {
		return err
	}
	return nil
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Agent exists
	found := false
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == opt.Name {
			found = true
			break
		}
	}
	if found == false {
		return exe, util.NewInputError(fmt.Sprintf("Could not find agent %s in the current namespace\n", opt.Name))
	}

	// Unmarshal file
	agentConfig := config.AgentConfiguration{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &agentConfig); err != nil {
		err = util.NewInputError("Could not unmarshall\n" + err.Error())
		return
	}

	return remoteExecutor{
		name:        opt.Name,
		agentConfig: agentConfig,
		namespace:   opt.Namespace,
	}, nil
}
