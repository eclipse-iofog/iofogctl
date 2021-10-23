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
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type microserviceExecutor struct {
	namespace string
	name      string
	filename  string
	client    *client.Client
	msvc      *client.MicroserviceInfo
}

func newMicroserviceExecutor(namespace, name, filename string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *microserviceExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	appName, msvcName, err := clientutil.ParseFQName(exe.name, "Microservice")
	if err != nil {
		return err
	}

	exe.msvc, err = exe.client.GetMicroserviceByName(appName, msvcName)
	return
}

func (exe *microserviceExecutor) GetName() string {
	return exe.name
}

func (exe *microserviceExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}

	if util.IsSystemMsvc(exe.msvc) {
		return nil
	}

	yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(exe.msvc, exe.client)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.MicroserviceKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: yamlMsvc,
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
