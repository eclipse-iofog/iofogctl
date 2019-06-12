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

package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	r := &remoteExecutor{}
	r.opt = opt
	return r
}

func (exe *remoteExecutor) Execute() (err error) {
	// Establish connection
	err = connect(exe.opt, exe.opt.Endpoint)
	if err != nil {
		return err
	}
	return config.Flush()
}
