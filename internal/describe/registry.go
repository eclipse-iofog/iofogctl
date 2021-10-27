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
	"fmt"
	"strconv"

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type registryExecutor struct {
	namespace string
	id        int
	filename  string
}

func newRegistryExecutor(namespace, name, filename string) (*registryExecutor, error) {
	a := &registryExecutor{}
	a.namespace = namespace
	id, err := strconv.Atoi(name)
	if err != nil {
		return nil, err
	}
	a.id = id
	a.filename = filename
	return a, nil
}

func (exe *registryExecutor) GetName() string {
	return strconv.Itoa(exe.id)
}

func (exe *registryExecutor) Execute() error {
	// Connect to controller
	ctrl, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	registriesList, err := ctrl.ListRegistries()
	if err != nil {
		return err
	}

	var registry rsc.Registry
	var private bool

	for _, r := range registriesList.Registries {
		if r.ID == exe.id {
			private = !r.IsPublic
			registry = rsc.Registry{
				URL:          &r.URL,
				ID:           r.ID,
				Private:      &private,
				Username:     &r.Username,
				Email:        &r.Email,
				Password:     nil,
				RequiresCert: &r.RequiresCert,
				Certificate:  &r.Certificate,
			}
			break
		}
	}

	if registry.ID == 0 {
		return util.NewNotFoundError(fmt.Sprintf("Could not find registry with ID %d", exe.id))
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RegistryKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
		},
		Spec: registry,
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
