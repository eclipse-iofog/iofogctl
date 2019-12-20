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

package deleteconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name      string
	namespace string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	return executor{name: name, namespace: namespace}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Detaching Connector")
	// Try to delete connector from Controller database
	ctrlClient, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	connectors, err := ctrlClient.ListConnectors()
	if err != nil {
		return err
	}
	for _, connector := range connectors.Connectors {
		if connector.Name == exe.name {
			if err = ctrlClient.DeleteConnector(connector.IP); err != nil {
				return err
			}

			// Try to detach from config
			// Ignore error, because only error is not found.
			config.DetachConnector(exe.namespace, exe.name)
			return config.Flush()
		}
	}

	return nil
}
