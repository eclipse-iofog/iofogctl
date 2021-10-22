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

package route

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func Execute(namespace, name, newName string) error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	route, err := clt.GetRoute(name)
	if err != nil {
		return err
	}

	if err := util.IsLowerAlphanumeric("Route", newName); err != nil {
		return err
	}
	util.SpinStart(fmt.Sprintf("Renaming route %s", name))
	route.Name = newName

	if err := clt.PatchRoute(name, &route); err != nil {
		return err
	}

	return err
}
