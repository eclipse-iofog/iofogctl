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

package get

type allExecutor struct {
	namespace string
}

func newAllExecutor(namespace string) *allExecutor {
	exe := &allExecutor{}
	exe.namespace = namespace
	return exe
}

func (exe *allExecutor) Execute() error {
	if err := newControllerExecutor(exe.namespace).Execute(); err != nil {
		return err
	}
	if err := newAgentExecutor(exe.namespace).Execute(); err != nil {
		return err
	}
	if err := newMicroserviceExecutor(exe.namespace).Execute(); err != nil {
		return err
	}

	return nil
}
