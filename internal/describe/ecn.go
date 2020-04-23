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

package describe

import "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"

type ecnExecutor struct {
	namespace   string
	name        string
	filename    string
	useDetached bool
	flow        *client.FlowInfo
	client      *client.Client
	msvcs       []client.MicroserviceInfo
	msvcPerID   map[string]*client.MicroserviceInfo
}

func newECNExecutor(namespace, name, filename string, useDetached bool) *ecnExecutor {
	a := &ecnExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	a.useDetached = useDetached
	return a
}

func (exe *ecnExecutor) GetName() string {
	return exe.name
}

func (exe *ecnExecutor) Execute() error {
	return nil
}
