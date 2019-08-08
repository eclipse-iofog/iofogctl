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

package deletemicroservice

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name string) error {
	// Disallow deletion of default
	if name == "default" {
		return util.NewInputError("Cannot delete default namespace")
	}

	// Get config
	ns, err := config.GetNamespace(name)
	if err != nil {
		return err
	}

	// Check resources exist
	hasAgents := len(ns.Agents) > 0
	hasControllers := len(ns.ControlPlane.Controllers) > 0
	hasMicroservices := len(ns.Microservices) > 0
	if hasAgents || hasControllers || hasMicroservices {
		return util.NewInputError("Namespace " + name + " not empty")
	}

	// Delete namespace
	err = config.DeleteNamespace(name)
	if err != nil {
		return err
	}

	return config.Flush()
}
