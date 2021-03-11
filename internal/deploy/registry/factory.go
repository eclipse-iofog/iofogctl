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

package deployregistry

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
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
	registry  rsc.Registry
}

func (exe *remoteExecutor) GetName() string {
	if exe.registry.URL != nil {
		return *exe.registry.URL
	}
	return ""
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying registry %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if exe.registry.ID > 0 {
		var publicPtr *bool
		if exe.registry.Private != nil {
			public := !*exe.registry.Private
			publicPtr = &public
		}
		return clt.UpdateRegistry(client.RegistryUpdateRequest{
			URL:          exe.registry.URL,
			IsPublic:     publicPtr,
			Certificate:  exe.registry.Certificate,
			RequiresCert: exe.registry.RequiresCert,
			Username:     exe.registry.Username,
			Email:        exe.registry.Email,
			Password:     exe.registry.Password,
			ID:           exe.registry.ID,
		})
	}

	createRequest := &client.RegistryCreateRequest{}
	if exe.registry.URL != nil {
		createRequest.URL = *exe.registry.URL
	}
	if exe.registry.Private != nil {
		createRequest.IsPublic = !*exe.registry.Private
	}
	if exe.registry.Certificate != nil {
		createRequest.Certificate = *exe.registry.Certificate
	}
	if exe.registry.RequiresCert != nil {
		createRequest.RequiresCert = *exe.registry.RequiresCert
	}
	if exe.registry.Username != nil {
		createRequest.Username = *exe.registry.Username
	}
	if exe.registry.Password != nil {
		createRequest.Password = *exe.registry.Password
	}
	if exe.registry.Email != nil {
		createRequest.Email = *exe.registry.Email
	}
	if _, err = clt.CreateRegistry(createRequest); err != nil {
		return err
	}

	return nil
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	// Unmarshal file
	var registry rsc.Registry
	if err = yaml.UnmarshalStrict(opt.Yaml, &registry); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if registry.Private == nil {
		Private := false
		registry.Private = &Private
	}

	if err := validate(registry, true); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		registry:  registry,
		namespace: opt.Namespace,
	}, nil
}

func validate(opt rsc.Registry, create bool) error {
	if create {
		if opt.URL == nil || *opt.URL == "" {
			return util.NewInputError("URL cannot be empty")
		}
		if opt.Email == nil || *opt.Email == "" {
			return util.NewInputError("Email cannot be empty")
		}
	}
	if opt.RequiresCert != nil && *opt.RequiresCert && opt.Certificate != nil && *opt.Certificate == "" {
		return util.NewInputError("Certificate cannot be empty if requiresCertificate is set to true")
	}
	if !*opt.Private && ((opt.Password == nil || *opt.Password == "") || (opt.Username == nil || *opt.Username == "")) {
		return util.NewInputError("Password and/or Username cannot be empty if Private is set to false")
	}

	return nil
}
