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

package connectcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace, name string, yaml []byte) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := unmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}

	// Kubernetes controller
	if controlPlane.Controllers[0].Kube.Config != "" {
		return newKubernetesExecutor(controlPlane, namespace), nil
	}

	// In YAML, the endpoint will come through Host variable
	controlPlane.Controllers[0].Endpoint = formatEndpoint(controlPlane.Controllers[0].SSH.Host)
	return newRemoteExecutor(controlPlane, namespace), nil
}

func NewManualExecutor(namespace, name, endpoint, kubeConfig, email, password string) (execute.Executor, error) {
	controlPlane := config.ControlPlane{
		IofogUser: config.IofogUser{
			Email:    email,
			Password: password,
		},
		Controllers: []config.Controller{
			{
				Name:     name,
				Endpoint: formatEndpoint(endpoint),
				Kube: config.Kube{
					Config: kubeConfig,
				},
			},
		},
	}

	if kubeConfig != "" {
		return newKubernetesExecutor(controlPlane, namespace), nil
	}

	// In manual approach, host address can be inferred from Endpoint variable
	controlPlane.Controllers[0].SSH.Host = util.Before(endpoint, ":")
	return newRemoteExecutor(controlPlane, namespace), nil
}

func formatEndpoint(endpoint string) string {
	before := util.Before(endpoint, ":")
	after := util.After(endpoint, ":")
	if after == "" {
		after = iofog.ControllerPortString
	}
	return before + ":" + after
}
