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
	"net/url"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"gopkg.in/yaml.v2"
)

type applicationTemplateExecutor struct {
	controller IofogController
	baseURL    *url.URL
	template   interface{}
	name       string
	client     *client.Client
}

func newApplicationTemplateExecutor(controller IofogController, controllerBaseURL *url.URL, template interface{}, name string) *applicationTemplateExecutor {
	exe := &applicationTemplateExecutor{
		controller: controller,
		baseURL:    controllerBaseURL,
		name:       name,
		template:   template,
	}

	return exe
}

func (exe *applicationTemplateExecutor) execute() error {
	// Init remote resources
	if err := exe.init(); err != nil {
		return err
	}

	// Deploy application
	return exe.deploy()
}

func (exe *applicationTemplateExecutor) init() (err error) {
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{BaseURL: exe.baseURL}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{BaseURL: exe.baseURL}, exe.controller.Email, exe.controller.Password)
	}

	return err
}

func (exe *applicationTemplateExecutor) deploy() error {
	file := IofogHeader{
		APIVersion: "iofog.org/v3",
		Kind:       ApplicationTemplateKind,
		Metadata: HeaderMetadata{
			Name: exe.name,
		},
		Spec: exe.template,
	}
	yamlBytes, err := yaml.Marshal(file)
	if err != nil {
		return err
	}
	existingAppTemplate, err := exe.client.GetApplicationTemplate(exe.name)
	// If not notfound error, return error
	if _, ok := err.(*client.NotFoundError); err != nil && !ok {
		return err
	}
	if existingAppTemplate == nil {
		if _, err := exe.client.CreateApplicationTemplateFromYAML(bytes.NewReader(yamlBytes)); err != nil {
			return err
		}
		return nil
	}
	if _, err := exe.client.UpdateApplicationTemplateFromYAML(exe.name, bytes.NewReader(yamlBytes)); err != nil {
		return err
	}
	return nil
}
