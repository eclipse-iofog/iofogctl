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

package deployvolumemount

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
	Name      string
}

type executor struct {
	volumeMount rsc.VolumeMount
	namespace   string
	name        string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) updateVolumeMount(clt *client.Client) (err error) {
	_, err = clt.GetVolumeMount(exe.name)
	if err != nil {
		return err
	}

	request := client.VolumeMountUpdateRequest{
		Name:          exe.name,
		ConfigMapName: exe.volumeMount.ConfigMapName,
		SecretName:    exe.volumeMount.SecretName,
	}

	if err = clt.UpdateVolumeMount(exe.name, &request); err != nil {
		return err
	}

	return nil
}

func (exe *executor) createVolumeMount(clt *client.Client) (err error) {
	request := client.VolumeMountCreateRequest{
		Name:          exe.name,
		ConfigMapName: exe.volumeMount.ConfigMapName,
		SecretName:    exe.volumeMount.SecretName,
	}

	if err = clt.CreateVolumeMount(&request); err != nil {
		return err
	}
	return nil
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying volume mount %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if _, err = clt.GetVolumeMount(exe.name); err != nil {
		return exe.createVolumeMount(clt)
	}
	return exe.updateVolumeMount(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Volume Mounts")
	}

	// Create volume mount
	volumeMount := rsc.VolumeMount{
		Name: opt.Name,
	}

	// Unmarshal spec
	if err = yaml.UnmarshalStrict(opt.Yaml, &volumeMount); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Validate input
	if opt.Name == "" {
		return nil, util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("Volume Mount", opt.Name); err != nil {
		return nil, err
	}

	return &executor{
		namespace:   opt.Namespace,
		volumeMount: volumeMount,
		name:        opt.Name,
	}, nil
}
