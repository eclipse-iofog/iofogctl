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

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	connectagent "github.com/eclipse-iofog/iofogctl/v2/internal/connect/agent"
	connectk8scontrolplane "github.com/eclipse-iofog/iofogctl/v2/internal/connect/controlplane/k8s"
	connectremotecontrolplane "github.com/eclipse-iofog/iofogctl/v2/internal/connect/controlplane/remote"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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

var kindOrder = []config.Kind{
	config.KubernetesControlPlaneKind,
	config.RemoteControlPlaneKind,
	config.LocalControlPlaneKind,
	config.KubernetesControllerKind,
	config.RemoteControllerKind,
	config.LocalControllerKind,
	config.AgentKind,
}

var kindHandlers = map[config.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	config.KubernetesControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return connectk8scontrolplane.NewExecutor(opt.Namespace, opt.Name, opt.YAML, config.KubernetesControlPlaneKind)
	},
	config.RemoteControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return connectremotecontrolplane.NewExecutor(opt.Namespace, opt.Name, opt.YAML, config.RemoteControlPlaneKind)
	},
	config.AgentKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return connectagent.NewExecutor(opt.Namespace, opt.Name, opt.YAML)
	},
}

func Execute(opt Options) error {
	// Check inputs
	if opt.InputFile != "" && (opt.ControllerEndpoint != "" || opt.KubeConfig != "") {
		return util.NewInputError("Either use a YAML file or provide Controller endpoint or Kube config to connect")
	}

	// TODO: refactor this to have less nesting
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
			if err == nil {
				if len(ns.Agents) != 0 || len(ns.GetControllers()) != 0 {
					return util.NewInputError("You must use an empty or non-existent namespace")
				}
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

		// K8s or Remote
		var exe execute.Executor
		if opt.KubeConfig != "" {
			exe, err = connectk8scontrolplane.NewManualExecutor(opt.Namespace, opt.ControllerName, opt.ControllerEndpoint, opt.KubeConfig, opt.IofogUserEmail, opt.IofogUserPass)
			if err != nil {
				return err
			}
		} else {
			// TODO: This doesn't make sense, connect to controlplane is passing in a controller name, it should be a list of controller details
			exe, err = connectremotecontrolplane.NewManualExecutor(opt.Namespace, opt.ControllerName, opt.ControllerEndpoint, opt.IofogUserEmail, opt.IofogUserPass)
			if err != nil {
				return err
			}
		}

		// Execute
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
