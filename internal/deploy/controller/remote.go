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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type remoteExecutor struct {
	opt *Options
}

func newRemoteExecutor(opt *Options) *remoteExecutor {
	d := &remoteExecutor{}
	d.opt = opt
	return d
}

func (exe *remoteExecutor) Execute() (err error) {
	// Instantiate installer
	installer := iofog.NewControllerInstaller(exe.opt.User, exe.opt.Host, exe.opt.Port, exe.opt.KeyFile)

	// Update configuration before we try to deploy in case of failure
	configEntry, err := prepareUserAndSaveConfig(exe.opt)
	if err != nil {
		return
	}

	// Install Controller and Connector
	if err = installer.Install(); err != nil {
		return
	}

	// Configure Controller and Connector
	if err = installer.Configure(iofog.User{
		Name:     configEntry.IofogUser.Name,
		Surname:  configEntry.IofogUser.Surname,
		Email:    configEntry.IofogUser.Email,
		Password: configEntry.IofogUser.Password,
	}); err != nil {
		return
	}

	// Update configuration
	configEntry.Endpoint = exe.opt.Host + ":54421" // TODO: change hardcode
	if err = config.UpdateController(exe.opt.Namespace, configEntry); err != nil {
		return
	}

	return config.Flush()
}
