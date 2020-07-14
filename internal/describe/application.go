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
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type applicationExecutor struct {
	namespace string
	name      string
	filename  string
	flow      *client.FlowInfo
	client    *client.Client
	msvcs     []client.MicroserviceInfo
	msvcPerID map[string]*client.MicroserviceInfo
}

func newApplicationExecutor(namespace, name, filename string) *applicationExecutor {
	a := &applicationExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}
	exe.flow, err = exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	msvcListResponse, err := exe.client.GetMicroservicesPerFlow(exe.flow.ID)
	if err != nil {
		return
	}

	// Filter system microservices
	for _, msvc := range msvcListResponse.Microservices {
		if util.IsSystemMsvc(msvc) {
			continue
		}
		exe.msvcs = append(exe.msvcs, msvc)
	}
	exe.msvcPerID = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(exe.msvcs); i++ {
		exe.msvcPerID[exe.msvcs[i].UUID] = &exe.msvcs[i]
	}
	return
}

func (exe *applicationExecutor) GetName() string {
	return exe.name
}

func (exe *applicationExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}

	yamlMsvcs := []rsc.Microservice{}
	yamlRoutes := []rsc.Route{}

	for _, msvc := range exe.msvcs {
		yamlMsvc, err := MapClientMicroserviceToDeployMicroservice(&msvc, exe.client)
		if err != nil {
			return err
		}
		// Remove fields
		yamlMsvc.Flow = nil
		yamlMsvcs = append(yamlMsvcs, *yamlMsvc)
	}

	application := rsc.Application{
		Name:          exe.flow.Name,
		Microservices: yamlMsvcs,
		Routes:        yamlRoutes,
		ID:            exe.flow.ID,
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ApplicationKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: application,
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
