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

package deleteagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name string) error {
	util.SpinStart("Deleting Agent")

	// Get executor
	exe, err := NewExecutor(namespace, name)
	if err != nil {
		return err
	}

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	// Update config
	if err = config.DeleteAgent(namespace, name); err != nil {
		return err
	}

	return config.Flush()
}
