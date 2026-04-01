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

package deployconfigmap

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
	Data      []byte
	Name      string
}

type executor struct {
	configMap rsc.ConfigMap
	namespace string
	name      string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) updateConfigMap(clt *client.Client) (err error) {
	_, err = clt.GetConfigMap(exe.name)
	if err != nil {
		return err
	}

	request := client.ConfigMapUpdateRequest{
		Name:      exe.name,
		Data:      exe.configMap.Data,
		Immutable: exe.configMap.Immutable,
	}

	if err = clt.UpdateConfigMap(exe.name, &request); err != nil {
		return err
	}

	return nil
}

func (exe *executor) createConfigMap(clt *client.Client) (err error) {
	request := client.ConfigMapCreateRequest{
		Name:      exe.name,
		Data:      exe.configMap.Data,
		Immutable: exe.configMap.Immutable,
	}

	if err = clt.CreateConfigMap(&request); err != nil {
		return err
	}
	return nil
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying configMap %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if _, err = clt.GetConfigMap(exe.name); err != nil {
		return exe.createConfigMap(clt)
	}
	return exe.updateConfigMap(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying ConfigMaps")
	}

	// Create secret
	configMap := rsc.ConfigMap{
		Name: opt.Name,
	}

	// Unmarshal spec (type)
	if err = yaml.UnmarshalStrict(opt.Yaml, &configMap); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Unmarshal data
	var data map[string]string
	if err = yaml.UnmarshalStrict(opt.Data, &data); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	configMap.Data = data

	// Validate input
	if opt.Name == "" {
		return nil, util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("ConfigMap", opt.Name); err != nil {
		return nil, err
	}

	return &executor{
		namespace: opt.Namespace,
		configMap: configMap,
		name:      opt.Name,
	}, nil
}
