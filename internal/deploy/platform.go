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
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deployconnector "github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

func Execute(opt *Options) error {
	// Check namespace exists
	if _, err := config.GetNamespace(opt.Namespace); err != nil {
		return err
	}

	// Read the input file to check validity of all resources before deploying any
	controlPlane, err := deploycontrolplane.UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}
	connectors, err := deployconnector.UnmarshallYAML(opt.InputFile)
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
	if len(controlPlane.Controllers) == 0 && len(connectors) == 0 && len(agents) == 0 && len(applications) == 0 {
		return util.NewInputError("No resources specified to deploy in the YAML file")
	}

	if len(controlPlane.Controllers) > 0 {
		// Require IofogUser
		if controlPlane.IofogUser.Email == "" || controlPlane.IofogUser.Name == "" || controlPlane.IofogUser.Password == "" || controlPlane.IofogUser.Surname == "" {
			return util.NewInputError("You must specify an ioFog user with a name, surname, email, and password")
		}
		// Deploy ControlPlane
		if err = deploycontrolplane.Execute(deploycontrolplane.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
			return err
		}
	}

	// Deploy Connectors
	if len(connectors) > 0 {
		if err = deployconnector.Execute(deployconnector.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
			return err
		}
	}

	// Deploy Agents
	if len(agents) > 0 {
		if err = deployagent.Execute(deployagent.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
			return err
		}
	}

	// Deploy Applications
	if len(applications) > 0 {
		if err = deployapplication.Execute(deployapplication.Options{Namespace: opt.Namespace, InputFile: opt.InputFile}); err != nil {
			return err
		}
	}

	return nil
}
