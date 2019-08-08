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

package deployagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Namespace string
	Name      string
	User      string
	Host      string
	Port      int
	KeyFile   string
	Local     bool
	Image     string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check Controller exists
	nbControllers := len(ns.ControlPlane.Controllers)
	if nbControllers != 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", nbControllers)
		return nil, util.NewInputError(errMessage)
	}

	// Local executor
	if opt.Local == true {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		exe, err := newLocalExecutor(opt, cli)
		if err != nil {
			return nil, err
		}
		return exe, nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
