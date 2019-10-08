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

package deployapplication

import (
	types "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
)

func Execute(controller types.IofogController, application types.Application) error {
	exe := newRemoteExecutor(controller, application)

	return exe.Execute()
}
