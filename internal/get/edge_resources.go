/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
)

type edgeResourceExecutor struct {
	namespace string
}

func newEdgeResourceExecutor(namespace string) *edgeResourceExecutor {
	return &edgeResourceExecutor{
		namespace: namespace,
	}
}

func (exe *edgeResourceExecutor) GetName() string {
	return ""
}

func (exe *edgeResourceExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateEdgeResourceOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateEdgeResourceOutput(namespace string) (table [][]string, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil && !rsc.IsNoControlPlaneError(err) {
		return
	}

	edgeResources := []client.EdgeResourceMetadata{}
	if err == nil {
		// Populate table
		listResponse, err := clt.ListEdgeResources()
		if err != nil {
			return table, err
		}
		edgeResources = listResponse.EdgeResources
	}

	return tabulateEdgeResources(edgeResources)
}

func tabulateEdgeResources(edgeResources []client.EdgeResourceMetadata) (table [][]string, err error) {
	// Generate table and headers
	table = make([][]string, len(edgeResources)+1)
	headers := []string{"EDGE RESOURCE", "PROTOCOL", "VERSIONS"}
	table[0] = append(table[0], headers...)

	// Coalesce versions
	index := make(map[string]client.EdgeResourceMetadata)
	for i := range edgeResources {
		edgeResource := edgeResources[i]
		name := edgeResource.Name
		if indexEdgeResource, exists := index[name]; exists {
			// Append version
			indexEdgeResource.Version = fmt.Sprintf("%s, %s", index[name].Version, edgeResource.Version)
			index[name] = indexEdgeResource
		} else {
			// Instantiate new resource
			index[name] = edgeResource
		}
	}
	// Populate rows
	idx := 0
	for i := range index {
		edge := index[i]
		// Store values
		row := []string{
			edge.Name,
			edge.InterfaceProtocol,
			edge.Version,
		}
		table[idx+1] = append(table[idx+1], row...)
		idx++
	}
	return table, err
}
