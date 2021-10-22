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

package deployapplicationtemplate

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	namespace string
	template  interface{}
	name      string
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying Application Template %s", exe.GetName()))

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return err
	}

	// Get Controller client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	controller := apps.IofogController{
		Endpoint: endpoint,
		Email:    controlPlane.GetUser().Email,
		Password: controlPlane.GetUser().Password,
		Token:    clt.GetAccessToken(),
	}
	baseURL, err := util.GetBaseURL(controller.Endpoint)
	if err != nil {
		return err
	}
	return apps.DeployApplicationTemplate(controller, baseURL, exe.template, exe.name)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	if _, err = config.GetNamespace(opt.Namespace); err != nil {
		return exe, err
	}
	// Unmarshal file
	var template interface{}
	if err = yaml.Unmarshal(opt.Yaml, &template); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	return &remoteExecutor{
		namespace: opt.Namespace,
		template:  template,
		name:      opt.Name,
	}, nil
}
