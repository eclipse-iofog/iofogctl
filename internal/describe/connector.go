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

import (
	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type connectorExecutor struct {
	namespace   string
	name        string
	filename    string
	useDetached bool
}

func newConnectorExecutor(namespace, name, filename string, useDetached bool) *connectorExecutor {
	c := &connectorExecutor{}
	c.namespace = namespace
	c.name = name
	c.filename = filename
	c.useDetached = useDetached
	return c
}

func (exe *connectorExecutor) GetName() string {
	return exe.name
}

func (exe *connectorExecutor) Execute() (err error) {
	var connector config.Connector
	if exe.useDetached {
		connector, err = config.GetDetachedConnector(exe.name)
	} else {
		connector, err = config.GetConnector(exe.namespace, exe.name)
	}
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: internal.LatestAPIVersion,
		Kind:       apps.ConnectorKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: connector,
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
