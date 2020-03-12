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

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

var clientByNamespace map[string]*client.Client = make(map[string]*client.Client)

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

	// If we are already authenticated, use existing client
	clt, ok := clientByNamespace[namespace]
	if !ok {
		clt, err = client.NewAndLogin(client.Options{Endpoint: endpoint}, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
		if err != nil {
			return
		}
		clientByNamespace[namespace] = clt
	}
	return
}

const APIVersionGroup = "iofog.org"
const LatestAPIVersion = APIVersionGroup + "/v2"

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

func MakeIntPtr(value int) *int {
	return &value
}

func MakeStrPtr(value string) *string {
	return &value
}

func MakeBoolPtr(value bool) *bool {
	return &value
}
