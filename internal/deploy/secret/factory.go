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

package deploysecret

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
	secret    rsc.Secret
	namespace string
	name      string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) updateSecret(clt *client.Client) (err error) {
	_, err = clt.GetSecret(exe.name)
	if err != nil {
		return err
	}

	request := client.SecretUpdateRequest{
		Name: exe.name,
		Data: exe.secret.Data,
	}

	if err = clt.UpdateSecret(exe.name, &request); err != nil {
		return err
	}

	return nil
}

func (exe *executor) createSecret(clt *client.Client) (err error) {
	request := client.SecretCreateRequest{
		Name: exe.name,
		Type: exe.secret.Type,
		Data: exe.secret.Data,
	}

	if err = clt.CreateSecret(&request); err != nil {
		return err
	}
	return nil
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying secret %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if _, err = clt.GetSecret(exe.name); err != nil {
		return exe.createSecret(clt)
	}
	return exe.updateSecret(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Secrets")
	}

	// Create secret
	secret := rsc.Secret{
		Name: opt.Name,
	}

	// Unmarshal spec (type)
	if err = yaml.UnmarshalStrict(opt.Yaml, &secret); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Unmarshal data
	var data map[string]string
	if err = yaml.UnmarshalStrict(opt.Data, &data); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}
	secret.Data = data

	// Validate input
	if opt.Name == "" {
		return nil, util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("Secret", opt.Name); err != nil {
		return nil, err
	}

	return &executor{
		namespace: opt.Namespace,
		secret:    secret,
		name:      opt.Name,
	}, nil
}
