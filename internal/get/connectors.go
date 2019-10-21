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
	namespace string
}

func newConnectorExecutor(namespace string) *connectorExecutor {
	a := &connectorExecutor{}
	a.namespace = namespace
	return a
}

func (exe *connectorExecutor) GetName() string {
	return ""
}

func (exe *connectorExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateConnectorOutput(exe.namespace); err != nil {
		return err
	}
	return config.Flush()
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
					Name: remoteConnector.Name,
					Host: remoteConnector.IP,
				}
				config.AddConnector(namespace, newConnectorConf)
			}

			clientConnectorIP := connectorsToPrint[remoteConnector.Name].Host
			if clientConnectorIP != remoteConnector.IP {
				util.PrintNotify("Detected endpoint discrepancy between client (" + clientConnectorIP + ") and server (" + remoteConnector.IP + ") for Connector " + remoteConnector.Name)
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
		"IP",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, connector := range connectors {
		// TODO: Ping Connector to check status
		// TODO: Get uptime
		// if UUID is empty, we assume the connector is not provided
		age, _ := util.ElapsedUTC(connector.Created, util.NowUTC())
		row := []string{
			connector.Name,
			"online",
			age,
			age,
			connector.Host,
		}
		table[idx+1] = append(table[idx+1], row...)
		idx = idx + 1
	}

	// Print table
	return print(table)
}
