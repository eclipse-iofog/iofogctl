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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type connectorExecutor struct {
	namespace string
	name      string
	filename  string
}

func newConnectorExecutor(namespace, name, filename string) *connectorExecutor {
	c := &connectorExecutor{}
	c.namespace = namespace
	c.name = name
	c.filename = filename
	return c
}

func (exe *connectorExecutor) GetName() string {
	return exe.name
}

func (exe *connectorExecutor) Execute() error {
	connector, err := config.GetConnector(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	header := deploy.Header{
		Kind: deploy.ConnectorKind,
		Metadata: deploy.HeaderMetadata{
			Namespace: exe.namespace,
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
