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
	"fmt"

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
	// Try to delete connector from Controller database
	ctrlClient, err := internal.NewControllerClient(exe.namespace)
	if err == nil {
		connectors, err := ctrlClient.ListConnectors()
		if err != nil {
			util.PrintInfo(fmt.Sprintf("Could not delete connector %s from the Controller. Error: %s\n", exe.name, err.Error()))
		} else {
			for _, connector := range connectors.Connectors {
				if connector.Name == exe.name {
					if err = ctrlClient.DeleteConnector(connector.IP); err != nil {
						util.PrintInfo(fmt.Sprintf("Could not delete connector %s from the Controller. Error: %s\n", exe.name, err.Error()))
					}
				}
			}
		}
	}

	// Try to remove iofog-connector
	connector, err := config.GetConnector(exe.namespace, exe.name)
	if err == nil {
		if util.IsLocalHost(connector.Host) {
			if err = exe.localRemove(); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-connector container. Error: %s\n", err.Error()))
			}
		} else if connector.Kube.Config != "" {
			if err = exe.k8sRemove(); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-connector from Kubernetes. Error: %s\n", err.Error()))
			}
		} else {
			if err = exe.remoteRemove(); err != nil {
				util.PrintInfo(fmt.Sprintf("Could not remove iofog-connector from the host %s. Error: %s\n", connector.Host, err.Error()))
			}
		}

		// Update config
		if err = config.DeleteConnector(exe.namespace, exe.name); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove iofog-agent from iofogctl config. Error: %s\n", err.Error()))
		} else {
			defer config.Flush()
		}
	} else {
		util.PrintInfo(fmt.Sprintf("Could not find iofog-connector in iofogctl configuration. Error: %s\n", err.Error()))
	}
	return nil
}
