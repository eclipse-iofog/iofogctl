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

type controlPlaneExecutor struct {
	namespace string
	filename  string
}

func newControlPlaneExecutor(namespace, filename string) *controlPlaneExecutor {
	c := &controlPlaneExecutor{}
	c.namespace = namespace
	c.filename = filename
	return c
}

func (exe *controlPlaneExecutor) GetName() string {
	return exe.namespace
}

func (exe *controlPlaneExecutor) Execute() error {
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	fmt.Printf("namespace: %s\n", exe.namespace)
	if exe.filename == "" {
		if err = util.Print(controlPlane); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(controlPlane, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
