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

package deploy

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	"github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	"github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

func Execute(opt *Options) error {
	// Check namespace option
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Read the input file to check validity of all resources before deploying any
	controlPlane, err := deploycontrolplane.UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}
	agents, err := deployagent.UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}
	applications, err := deployapplication.UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}
	// If there are no resources return error
	if len(controlPlane.Controllers) == 0 && len(agents) == 0 && len(applications) == 0 {
		return util.NewInputError("No resources specified to deploy in the YAML file")
	}
	// If no controller is provided, one must already exist
	if len(controlPlane.Controllers) == 0 {
		if len(ns.ControlPlane.Controllers) == 0 {
			return util.NewInputError("If you are not deploying a new controller, one must exist in the specified namespace")
		}
	}

	// Deploy ControlPlane
	if err = deploycontrolplane.Execute(deploycontrolplane.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
		return err
	}

	// Deploy Agents
	if err = deployagent.Execute(deployagent.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
		return err
	}

	// Deploy Applications
	if err = deployapplication.Execute(deployapplication.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
		return err
	}

	return nil
}
