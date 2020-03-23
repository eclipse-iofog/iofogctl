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

package controller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, newName string) error {
	// Check that Controller exists in current namespace
	ctrl, err := config.GetController(namespace, name)
	if err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming Controller %s", name))
	ctrl.SetName(newName)
	if err = config.UpdateController(namespace, ctrl); err != nil {
		return err
	}
	if err = config.DeleteController(namespace, name); err != nil {
		return err
	}
	config.Flush()
	return nil
}
