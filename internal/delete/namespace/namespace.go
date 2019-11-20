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
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/all"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name string, force bool) error {
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

	// Force must be specified
	if !force && (hasAgents || hasControllers) {
		return util.NewInputError("Namespace " + name + " not empty. You must force the deletion if the namespace is not empty")
	}

	// Handle delete all
	if force && (hasAgents || hasControllers) {
		if err = delete.Execute(name); err != nil {
			return err
		}
	}

	// Delete namespace
	err = config.DeleteNamespace(name)
	if err != nil {
		return err
	}

	return config.Flush()
}
