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
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type connectorExecutor struct {
	namespace    string
	showDetached bool
}

func newConnectorExecutor(namespace string, showDetached bool) *connectorExecutor {
	a := &connectorExecutor{}
	a.namespace = namespace
	a.showDetached = showDetached
	return a
}

func (exe *connectorExecutor) GetName() string {
	return ""
}

func (exe *connectorExecutor) Execute() error {
	if exe.showDetached {
		printDetached()
		if err := generateDetachedConnectorOutput(); err != nil {
			return err
		}
		return nil
	}
	printNamespace(exe.namespace)
	if err := generateConnectorOutput(exe.namespace); err != nil {
		return err
	}
	return config.Flush()
}

func generateDetachedConnectorOutput() error {
	detachedResources := config.GetDetachedResources()
	connectors := []config.Connector{}
	for _, connector := range detachedResources.Connectors {
		connectors = append(connectors, connector)
	}
	return tabulateConnectors(connectors)
}

func generateConnectorOutput(namespace string) error {
	// Get Config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Make an index of Connectors to print
	connectorsToPrint := make(map[string]config.Connector)
	for _, connector := range ns.Connectors {
		connectorsToPrint[connector.Name] = connector
	}

	// Connect to Controller if it is ready
	endpoint, err := ns.ControlPlane.GetControllerEndpoint()
	if err == nil {
		// Instantiate client
		// Log into Controller
		ctrlClient, err := client.NewAndLogin(endpoint, ns.ControlPlane.IofogUser.Email, ns.ControlPlane.IofogUser.Password)
		if err != nil {
			return tabulateConnectors(ns.Connectors)
		}

		// Get Connectors from Controller
		listConnectorsResponse, err := ctrlClient.ListConnectors()
		if err != nil {
			return err
		}

		// Process Connectors
		for _, remoteConnector := range listConnectorsResponse.Connectors {
			// Server may have connectors that the client is not aware of, update config if so
			if _, exists := connectorsToPrint[remoteConnector.Name]; !exists {
				newConnectorConf := config.Connector{
					Name:     remoteConnector.Name,
					Endpoint: remoteConnector.IP,
				}
				config.AddConnector(namespace, newConnectorConf)
			}
		}
	}

	return tabulateConnectors(ns.Connectors)
}

func tabulateConnectors(connectors []config.Connector) error {
	// Generate table and headers
	table := make([][]string, len(connectors)+1)
	headers := []string{
		"CONNECTOR",
		"STATUS",
		"AGE",
		"UPTIME",
		"ADDR",
		"PORT",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, connector := range connectors {
		// TODO: Ping Connector to check status
		// TODO: Get uptime
		// if UUID is empty, we assume the connector is not provided
		age, _ := util.ElapsedUTC(connector.Created, util.NowUTC())
		endpoint, port := getEndpointAndPort(connector.Endpoint, "8080")
		row := []string{
			connector.Name,
			"online",
			age,
			age,
			endpoint,
			port,
		}
		table[idx+1] = append(table[idx+1], row...)
		idx = idx + 1
	}

	// Print table
	return print(table)
}
