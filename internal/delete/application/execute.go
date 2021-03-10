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

package deleteapplication

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
)

func Execute(namespace, name string) error {
	// Get executor
	exe, _ := NewExecutor(namespace, name)

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	// Leave this here as a note on general practice with Execute functions
	return config.Flush()
}
