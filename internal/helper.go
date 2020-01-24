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

package internal

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// NewControllerClient returns an iofog-go-sdk/client configured for the current namespace
func NewControllerClient(namespace string) (clt *client.Client, err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil {
		return nil, err
	}
	endpoint, err := controlPlane.GetControllerEndpoint()
	if err != nil {
		return nil, err
	}
	return client.NewAndLogin(endpoint, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
}

const APIVersionGroup = "iofog.org"
const LatestAPIVersion = APIVersionGroup + "/v1"

var supportedAPIVersionsMap = map[string]bool{
	LatestAPIVersion: true,
}

func ValidateHeader(header config.Header) error {
	if _, found := supportedAPIVersionsMap[header.APIVersion]; found == false {
		return util.NewInputError(fmt.Sprintf("Unsupported YAML API version %s.\nPlease use version %s. See iofog.org for specification details.", header.APIVersion, LatestAPIVersion))
	}
	return nil
}

func IsSystemAgent(agentConfig config.AgentConfiguration) bool {
	return agentConfig.IsSystem != nil && *agentConfig.IsSystem
}
