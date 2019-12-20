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

package attachconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name        string
	Namespace   string
	Host        string
	User        string
	Port        int
	KeyFile     string
	KubeConfig  string
	UseDetached bool
}

type executor struct {
	opt  Options
	cnct config.Connector
}

func NewExecutor(opt Options) (execute.Executor, error) {
	return executor{opt: opt}, nil
}

func (exe executor) GetName() string {
	return exe.opt.Name
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Connector")
	var err error
	if exe.opt.UseDetached {
		exe.cnct, err = config.GetDetachedConnector(exe.opt.Name)
	} else {
		exe.cnct = config.Connector{
			Name: exe.opt.Name,
			Host: exe.opt.Host,
			Kube: config.Kube{
				Config: exe.opt.KubeConfig,
			},
			SSH: config.SSH{
				User:    exe.opt.User,
				KeyFile: exe.opt.KeyFile,
				Port:    exe.opt.Port,
			},
		}
	}

	if err != nil {
		return err
	}

	if util.IsLocalHost(exe.cnct.Host) {
		err = exe.localAttach()
	} else if exe.cnct.Kube.Config != "" {
		err = exe.k8sAttach()
	} else {
		err = exe.remoteAttach()
	}

	if err != nil {
		return err
	}

	if exe.opt.UseDetached {
		if err = config.AttachConnector(exe.opt.Namespace, exe.opt.Name); err != nil {
			return err
		}
	} else {
		if err = config.UpdateConnector(exe.opt.Namespace, exe.cnct); err != nil {
			return err
		}
	}

	return config.Flush()
}
