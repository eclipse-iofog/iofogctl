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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type microservice struct {
}

func New() *microservice {
	c := &microservice{}
	return c
}

func (ctrl *microservice) Execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := config.DeleteMicroservice(namespace, name)
	if err != nil {
		return err
	}

	fmt.Printf("\nMicroservice %s/%s successfully deleted.\n", namespace, name)

	return config.Flush()
}
