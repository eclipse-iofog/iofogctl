/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package deployservice

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Tags      *[]string
	Name      string
}

type executor struct {
	service   rsc.ClusterService
	namespace string
	name      string
	tags      []string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) updateService(clt *client.Client) (err error) {
	_, err = clt.GetService(exe.name)
	if err != nil {
		return err
	}

	request := client.ServiceUpdateRequest{
		Name:          exe.name,
		Type:          exe.service.Type,
		Resource:      exe.service.Resource,
		TargetPort:    exe.service.TargetPort,
		ServicePort:   exe.service.ServicePort,
		K8sType:       exe.service.K8sType,
		DefaultBridge: exe.service.DefaultBridge,
		Tags:          exe.tags,
	}

	if err = clt.UpdateService(exe.name, &request); err != nil {
		return err
	}

	return nil
}

func (exe *executor) createService(clt *client.Client) (err error) {
	request := client.ServiceCreateRequest{
		Name:          exe.name,
		Type:          exe.service.Type,
		Resource:      exe.service.Resource,
		TargetPort:    exe.service.TargetPort,
		ServicePort:   exe.service.ServicePort,
		K8sType:       exe.service.K8sType,
		DefaultBridge: exe.service.DefaultBridge,
		Tags:          exe.tags,
	}

	if err = clt.CreateService(&request); err != nil {
		return err
	}
	return nil
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying service %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if _, err = clt.GetService(exe.name); err != nil {
		return exe.createService(clt)
	}
	return exe.updateService(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Services")
	}

	// Create service
	service := rsc.ClusterService{
		Name: opt.Name,
	}

	// Unmarshal spec
	if err = yaml.UnmarshalStrict(opt.Yaml, &service); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Validate input
	if opt.Name == "" {
		return nil, util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("Service", opt.Name); err != nil {
		return nil, err
	}

	// Handle tags
	var tags []string
	if opt.Tags != nil {
		tags = *opt.Tags
	}

	return &executor{
		namespace: opt.Namespace,
		service:   service,
		name:      opt.Name,
		tags:      tags,
	}, nil
}
