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

package apps

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"gopkg.in/yaml.v2"
)

type applicationExecutor struct {
	controller      IofogController
	app             interface{}
	name            string
	applicationInfo *client.ApplicationInfo
	client          *client.Client
}

func newApplicationExecutor(controller IofogController, app interface{}, name string) *applicationExecutor {
	exe := &applicationExecutor{
		controller: controller,
		app:        app,
		name:       name,
	}

	return exe
}

func (exe *applicationExecutor) execute() (err error) {
	// Init remote resources
	if err := exe.init(); err != nil {
		return err
	}

	// Try application API
	// Look for exisiting application
	exe.applicationInfo, err = exe.client.GetApplicationByName(exe.name)

	// If not notfound error, return error
	if _, ok := err.(*client.NotFoundError); err != nil && !ok {
		return err
	}

	// Deploy application
	if err := exe.deploy(); err != nil {
		return err
	}
	return nil
}

func (exe *applicationExecutor) init() (err error) {
	baseURL, err := url.Parse(exe.controller.Endpoint)
	if err != nil {
		return fmt.Errorf(errParseControllerURL, err.Error())
	}
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{BaseURL: baseURL}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{BaseURL: baseURL}, exe.controller.Email, exe.controller.Password)
	}
	return err
}

func (exe *applicationExecutor) create() (err error) {
	file := IofogHeader{
		APIVersion: "iofog.org/v3",
		Kind:       ApplicationKind,
		Metadata: HeaderMetadata{
			Name: exe.name,
		},
		Spec: exe.app,
	}
	yamlBytes, err := yaml.Marshal(file)
	if err != nil {
		return err
	}
	if _, err = exe.client.CreateApplicationFromYAML(bytes.NewReader(yamlBytes)); err != nil {
		return err
	}
	return nil
}

func (exe *applicationExecutor) update() (err error) {
	file := IofogHeader{
		APIVersion: "iofog.org/v3",
		Kind:       ApplicationKind,
		Metadata: HeaderMetadata{
			Name: exe.name,
		},
		Spec: exe.app,
	}
	yamlBytes, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	if _, err = exe.client.UpdateApplicationFromYAML(exe.name, bytes.NewReader(yamlBytes)); err != nil {
		return err
	}
	return nil
}

func (exe *applicationExecutor) deploy() (err error) {
	// Existing app info retrieved in init
	if exe.applicationInfo == nil {
		if err := exe.create(); err != nil {
			return err
		}
	} else {
		if err := exe.update(); err != nil {
			return err
		}
	}

	// Start application
	if _, err = exe.client.StartApplication(exe.name); err != nil {
		return err
	}
	return nil
}
