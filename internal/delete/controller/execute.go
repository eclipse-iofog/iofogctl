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

package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func Execute(namespace, name string) error {
	util.SpinStart("Deleting Controller")

	// Get executor
	exe, err := NewExecutor(namespace, name)
	if err != nil {
		return err
	}

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	return config.Flush()
}
