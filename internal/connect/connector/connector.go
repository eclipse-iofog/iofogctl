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

package connectconnector

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	connector config.Connector
	namespace string
}

func (exe executor) GetName() string {
	return exe.connector.Name
}

func (exe executor) Execute() error {
	connectors, err := config.GetConnectors(exe.namespace)
	if err != nil {
		return err
	}

	for _, connector := range connectors {
		if connector.Name == exe.connector.Name {
			// Only update ssh info
			connector.SSH.KeyFile = exe.connector.SSH.KeyFile
			connector.SSH.Port = exe.connector.SSH.Port
			connector.SSH.User = exe.connector.SSH.User
			config.UpdateConnector(exe.namespace, connector)
			return nil
		}
	}

	util.PrintNotify(fmt.Sprintf("ECN does not contain connector %s\n", exe.connector.Name))
	return nil
}

func NewExecutor(namespace, name string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	connector, err := unmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}
	connector.Name = name

	return executor{namespace: namespace, connector: connector}, nil
}
