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

package describe

import (
	apps "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	exe.client, err = internal.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}
	exe.msvc, err = exe.client.GetMicroserviceByName(exe.name)
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

	if util.IsSystemMsvc(*(exe.msvc)) {
		return nil
	}

	yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(exe.msvc, exe.client)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: internal.LatestAPIVersion,
		Kind:       config.Kind(apps.MicroserviceKind),
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: yamlMsvc,
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
