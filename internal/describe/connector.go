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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
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
	fmt.Printf("namespace: %s\n", exe.namespace)
	if exe.filename == "" {
		if err = util.Print(connector); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(connector, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
