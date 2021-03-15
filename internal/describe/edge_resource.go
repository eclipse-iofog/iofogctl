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

package describe

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type edgeResourceExecutor struct {
	namespace string
	name      string
	version   string
	filename  string
}

func newEdgeResourceExecutor(namespace, name, version, filename string) *edgeResourceExecutor {
	return &edgeResourceExecutor{
		namespace: namespace,
		name:      name,
		version:   version,
		filename:  filename,
	}
}

func (exe *edgeResourceExecutor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.name, exe.version)
}

func (exe *edgeResourceExecutor) Execute() error {
	_, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Edge Resource
	edge, err := clt.GetHTTPEdgeResourceByName(exe.name, exe.version)
	if err != nil {
		return err
	}

	// Convert to YAML
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.EdgeResourceKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: rsc.EdgeResource{
			Description:       edge.Description,
			Display:           edge.Display,
			Interface:         &edge.Interface,
			InterfaceProtocol: edge.InterfaceProtocol,
			Name:              edge.Name,
			OrchestrationTags: edge.OrchestrationTags,
			Version:           edge.Version,
			Custom:            edge.Custom,
		},
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
