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
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (agent config.Agent, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &agent); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	// None specified
	if agent.Host == "" {
		return
	}

	if agent.Port == 0 {
		agent.Port = 22
	}
	// Format file paths
	if agent.KeyFile, err = util.FormatPath(agent.KeyFile); err != nil {
		return
	}

	return
}

func Validate(agent config.Agent) error {
	if agent.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Agents")
	}
	if (agent.Host != "localhost" && agent.Host != "127.0.0.1") && (agent.Host == "" || agent.User == "" || agent.KeyFile == "") {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return nil
}
