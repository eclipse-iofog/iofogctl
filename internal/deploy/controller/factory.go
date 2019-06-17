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
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

type Options struct {
	Name             string
	Namespace        string
	User             string
	Host             string
	KeyFile          string
	Local            bool
	KubeConfig       string
	KubeControllerIP string
	ImagesFile       string
	Images           map[string]string
}

func NewExecutor(opt *Options) (Executor, error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	// Check the namespace does not contain a Controller yet
	nbControllers := len(ns.Controllers)
	if nbControllers > 0 {
		return nil, util.NewInputError("This namespace already contains a Controller. Please remove it before deploying a new one.")
	}

	// Local executor
	if opt.Local == true {
		cli, err := iofog.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(opt, cli), nil
	}

	// Kubernetes executor
	if opt.KubeConfig != "" {
		// If image file specified, read it
		if opt.ImagesFile != "" {
			opt.Images = make(map[string]string)
			err := util.UnmarshalYAML(opt.ImagesFile, opt.Images)
			if err != nil {
				return nil, err
			}
		}
		return newKubernetesExecutor(opt), nil
	}

	// Default executor
	if opt.Host == "" || opt.KeyFile == "" || opt.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newRemoteExecutor(opt), nil
}
