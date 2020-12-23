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

package get

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
)

const (
	separateDelim = ", "
)

type applicationTemplateExecutor struct {
	namespace string
	templates []client.ApplicationTemplate
}

func newApplicationTemplateExecutor(namespace string) *applicationTemplateExecutor {
	c := &applicationTemplateExecutor{}
	c.namespace = namespace
	return c
}

func (exe *applicationTemplateExecutor) GetName() string {
	return ""
}

func (exe *applicationTemplateExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	table := exe.generateApplicationTemplateOutput()
	return print(table)
}

func (exe *applicationTemplateExecutor) init() (err error) {
	// Init client
	clt, err := iutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}

	// Get templates from Controller
	resp, err := clt.ListApplicationTemplates()
	if err != nil {
		return err
	}
	exe.templates = resp.ApplicationTemplates

	return
}

func (exe *applicationTemplateExecutor) generateApplicationTemplateOutput() (table [][]string) {
	// Generate table and headers
	table = make([][]string, len(exe.templates)+1)
	headers := []string{"TEMPLATE", "DESCRIPTION", "MICROSERVICES", "ROUTES"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, template := range exe.templates {
		row := []string{
			template.Name,
			template.Description,
			encodeMicroservices(template.Application.Microservices),
			encodeRoutes(template.Application.Routes),
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return
}

func encodeMicroservices(msvcs []client.MicroserviceCreateRequest) (encoded string) {
	for idx := range msvcs {
		msvc := &msvcs[idx]
		delim := separateDelim
		if idx == 0 {
			delim = ""
		}
		encoded = fmt.Sprintf("%s%s%s", encoded, delim, msvc.Name)
	}
	return
}

func encodeRoutes(routes []client.ApplicationRouteCreateRequest) (encoded string) {
	for routeIdx, route := range routes {
		delim := separateDelim
		if routeIdx == 0 {
			delim = ""
		}
		encoded = fmt.Sprintf("%s%s%s", encoded, delim, route.Name)
	}
	return
}
