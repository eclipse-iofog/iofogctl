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
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	connectagent "github.com/eclipse-iofog/iofogctl/internal/connect/agent"
	connectconnector "github.com/eclipse-iofog/iofogctl/internal/connect/connector"
	connectcontroller "github.com/eclipse-iofog/iofogctl/internal/connect/controller"
	connectcontrolplane "github.com/eclipse-iofog/iofogctl/internal/connect/controlplane"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace          string
	OverwriteNamespace bool
	InputFile          string
	ControllerName     string
	ControllerEndpoint string
	KubeConfig         string
	IofogUserEmail     string
	IofogUserPass      string
}

var kindOrder = []apps.Kind{
	apps.ControlPlaneKind,
	apps.ControllerKind,
	apps.ConnectorKind,
	apps.AgentKind,
}

var kindHandlers = map[apps.Kind]func(string, string, []byte) (execute.Executor, error){
	apps.ControlPlaneKind: connectcontrolplane.NewExecutor,
	apps.AgentKind:        connectagent.NewExecutor,
	apps.ConnectorKind:    connectconnector.NewExecutor,
	apps.ControllerKind:   connectcontroller.NewExecutor,
}

func Execute(opt Options) error {
	// Check inputs
	if opt.InputFile != "" && hasFlags(opt) {
		return util.NewInputError("Either use a YAML file or provide Controller endpoint or kube config along with a Controller name and ioFog user email/password.")
	}

	// Check for existing namespace
	ns, err := config.GetNamespace(opt.Namespace)
	if err == nil {
		// Overwrite namespace if requested
		if opt.OverwriteNamespace {
			delErr := config.DeleteNamespace(opt.Namespace)
			addErr := config.AddNamespace(opt.Namespace, util.NowUTC())
			if delErr != nil || addErr != nil {
				return util.NewInternalError("Failed to overwrite namespace " + opt.Namespace)
			}
		} else {
			// Check the namespace is empty
			if len(ns.Agents) != 0 || len(ns.ControlPlane.Controllers) != 0 {
				return util.NewInputError("You must use an empty or non-existent namespace")
			}
		}
	} else {
		// Create namespace
		if err = config.AddNamespace(opt.Namespace, util.NowUTC()); err != nil {
			return err
		}
	}
	// Flush at the end
	defer config.Flush()

	if opt.InputFile != "" {
		if err = executeWithYAML(opt.InputFile, opt.Namespace); err != nil {
			return err
		}
	} else {
		if !hasAllFlags(opt) {
			return util.NewInputError("If no YAML file is provided, must provide Controller endpoint or kube config along with Controller name and ioFog user email/password")
		}
		exe, err := connectcontrolplane.NewManualExecutor(opt.Namespace, opt.ControllerName, opt.ControllerEndpoint, opt.KubeConfig, opt.IofogUserEmail, opt.IofogUserPass)
		if err != nil {
			return err
		}
		if err = exe.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func executeWithYAML(yamlFile, namespace string) error {
	executorsMap, err := execute.GetExecutorsFromYAML(yamlFile, namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Controlplane, Controller, Connector, Agent
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("connect %s", kindOrder[idx])); err != nil {
			return err
		}
	}

	return nil
}

func hasAllFlags(opt Options) bool {
	return opt.ControllerName != "" && opt.IofogUserEmail != "" && opt.IofogUserPass != "" && (opt.KubeConfig != "" || opt.ControllerEndpoint != "")
}

func hasFlags(opt Options) bool {
	return opt.KubeConfig != "" || opt.ControllerEndpoint != "" || opt.ControllerName != "" || opt.IofogUserEmail != "" || opt.IofogUserPass != ""
}
