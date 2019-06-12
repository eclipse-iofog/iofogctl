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
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type microserviceExecutor struct {
	namespace string
	name      string
}

func newMicroserviceExecutor(namespace, name string) *microserviceExecutor {
	m := &microserviceExecutor{}
	m.namespace = namespace
	m.name = name
	return m
}

func (ms *microserviceExecutor) Execute() error {
	microservice, err := config.GetMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}
	if err = print(microservice); err != nil {
		return err
	}
	return nil
}
