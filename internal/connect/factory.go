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
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Name      string
	Endpoint  string
	KubeFile  string
	Email     string
	Password  string
}

type Executor interface {
	Execute() error
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check namespace is empty
	ns, err := config.GetNamespace(opt.Namespace)
	if err == nil {
		// Namespace exists, must be empty
		if len(ns.Agents) != 0 || len(ns.Controllers) != 0 || len(ns.Microservices) != 0 {
			return nil, util.NewInputError("You must use an empty or non-existent namespace")
		}
	} else {
		// Create namespace
		if err = config.AddNamespace(opt.Namespace, util.NowUTC()); err != nil {
			return nil, err
		}
	}

	// User details
	if opt.Email == "" || opt.Password == "" {
		return nil, util.NewInputError("You must specify email and password of user registered against the Controller")
	}

	// Kubernetes controller
	if opt.KubeFile != "" {
		return newKubernetesExecutor(opt), nil
	}

	// Remote controller needs host address
	if opt.Endpoint == "" {
		return nil, util.NewInputError("Must specify Controller host and port if connecting to non-Kubernetes Controller")
	}
	return newRemoteExecutor(opt), nil
}
