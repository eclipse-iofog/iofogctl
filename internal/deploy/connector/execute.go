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

package deployconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	InputFile string
}

func Execute(opt Options) error {
	// Make sure to update config despite failure
	defer config.Flush()

	// Check the namespace exists
	_, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Unmarshall file
	connectors, err := UnmarshallYAML(opt.InputFile)
	if err != nil {
		return err
	}

	// Should be atleast one to deploy
	if len(connectors) == 0 {
		return util.NewError("Could not read any Connectors from YAML")
	}

	// Output message
	msg := "Deploying Connector"
	if len(connectors) > 1 {
		msg += "s"
	}
	util.SpinStart(msg)

	// Get the Control Plane
	controlPlane, err := config.GetControlPlane(opt.Namespace)
	if err != nil {
		return err
	}

	// Execute Connectors
	var executors []execute.Executor
	for idx := range connectors {
		exe, err := NewExecutor(opt.Namespace, &connectors[idx], controlPlane)
		if err != nil {
			return err
		}
		// TODO: Replace serial execution when CRD updated
		// Kubernetes Connectors cannot run in parallel
		if connectors[idx].KubeConfig != "" {
			if err = exe.Execute(); err != nil {
				return err
			}
		} else {
			executors = append(executors, exe)
		}
	}

	// Run parallel executors
	if err := runExecutors(executors); err != nil {
		return err
	}

	// Update configuration
	for idx := range connectors {
		if err = config.UpdateConnector(opt.Namespace, connectors[idx]); err != nil {
			return err
		}
	}

	return nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}
	return nil
}
