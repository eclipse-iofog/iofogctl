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
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type edgeResourceExecutor struct {
	namespace   string
	nameVersion string
	filename    string
}

func newEdgeResourceExecutor(namespace, nameVersion, filename string) *edgeResourceExecutor {
	return &edgeResourceExecutor{
		namespace:   namespace,
		nameVersion: nameVersion,
		filename:    filename,
	}
}

func (exe *edgeResourceExecutor) GetName() string {
	return exe.nameVersion
}

func (exe *edgeResourceExecutor) Execute() error {
	_, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Decode nameVersion
	name, version, err := iutil.DecodeNameVersion(exe.nameVersion)
	if err != nil {
		return err
	}

	// Connect to Controller
	clt, err := iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Edge Resource
	edge, err := clt.GetHttpEdgeResourceByName(name, version)
	if err != nil {
		return err
	}

	// Convert to YAML
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.EdgeResourceKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      name,
		},
		Spec: rsc.EdgeResource{
			Description:       edge.Description,
			Display:           edge.Display,
			Interface:         &edge.Interface,
			InterfaceProtocol: edge.InterfaceProtocol,
			Name:              edge.Name,
			OrchestrationTags: edge.OrchestrationTags,
			Version:           edge.Version,
		},
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
