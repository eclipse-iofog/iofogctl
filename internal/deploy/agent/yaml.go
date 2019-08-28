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

package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type specification struct {
	Agents []config.Agent
}

func UnmarshallYAML(filename string) (agents []config.Agent, err error) {
	// Unmarshall the input file
	var spec specification
	if err = util.UnmarshalYAML(filename, &spec); err != nil || len(spec.Agents) == 0 {
		var agent config.Agent
		if err = util.UnmarshalYAML(filename, &agent); err != nil {
			err = util.NewInputError("Could not unmarshall " + filename + "\n" + err.Error())
			return
		}
		// None specified
		if agent.Host == "" {
			return
		}
		//Validate
		if err = validate(agent); err != nil {
			return
		}
		// Append the single agent
		agents = append(agents, agent)
	} else {
		// Record multiple agents
		agents = spec.Agents
	}

	for idx := range agents {
		agent := &agents[idx]
		// Fix SSH port
		if agent.Port == 0 {
			agent.Port = 22
		}
		// Format file paths
		if agent.KeyFile, err = util.FormatPath(agent.KeyFile); err != nil {
			return
		}
	}

	return
}

func validate(agent config.Agent) error {
	if agent.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Agents")
	}
	if agent.Host == "" || agent.User == "" || agent.KeyFile == "" {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
