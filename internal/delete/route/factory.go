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

package deleteroute

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
)

type executor struct {
	namespace string
	name      string
}

func (exe executor) GetName() string {
	return "deleting Route " + exe.name
}

func (exe executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	appName, routeName, err := clientutil.ParseFQName(exe.name, "Route")
	if err != nil {
		return err
	}

	if err = clt.DeleteRoute(appName, routeName); err != nil {
		return
	}
	return
}

func NewExecutor(namespace, name string) (exe execute.Executor) {
	return executor{
		namespace: namespace,
		name:      name,
	}
}
