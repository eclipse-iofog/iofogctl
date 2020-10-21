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

func (exe *Executor) initLegacy() (err error) {
	flow, err := exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	exe.flow = flow
	return
}

func (exe *Executor) deleteLegacy() (err error) {
	// Init remote resources
	if err = exe.initLegacy(); err != nil {
		return
	}

	// Delete flow
	if err = exe.client.DeleteFlow(exe.flow.ID); err != nil {
		return
	}
	return
}
