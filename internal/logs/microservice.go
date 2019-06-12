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

package logs

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

func (ns *microserviceExecutor) Execute() error {
	return nil
}
