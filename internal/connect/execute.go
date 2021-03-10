/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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
	"encoding/base64"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	connectk8scontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/connect/controlplane/k8s"
	connectremotecontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/connect/controlplane/remote"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
	Generate           bool
	Base64Encoded      bool
}

var kindOrder = []config.Kind{
	config.KubernetesControlPlaneKind,
	config.RemoteControlPlaneKind,
}

var kindHandlers = map[config.Kind]func(*execute.KindHandlerOpt) (execute.Executor, error){
	config.KubernetesControlPlaneKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return connectk8scontrolplane.NewExecutor(opt.Namespace, opt.Name, opt.YAML, config.KubernetesControlPlaneKind)
	},
	config.RemoteControlPlaneKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return connectremotecontrolplane.NewExecutor(opt.Namespace, opt.Name, opt.YAML, config.RemoteControlPlaneKind)
	},
}

func Execute(opt *Options) error {
	if opt.Generate {
		return generateConnectionString(opt.Namespace)
	}

	if !opt.Base64Encoded {
		opt.IofogUserPass = base64.StdEncoding.EncodeToString([]byte(opt.IofogUserPass))
	}

	// Check inputs
	if opt.InputFile != "" && (opt.ControllerEndpoint != "" || opt.KubeConfig != "") {
		return util.NewInputError("Either use a YAML file or provide Controller endpoint or Kube config to connect")
	}

	// Check for existing namespace
	ns, err := config.GetNamespace(opt.Namespace)
	if err == nil {
		// Check the namespace is empty
		if len(ns.GetAgents()) != 0 || len(ns.GetControllers()) != 0 {
			if !opt.OverwriteNamespace {
				return util.NewInputError("You must use an empty or non-existent namespace")
			}
			// Overwrite
			delErr := config.DeleteNamespace(opt.Namespace)
			addErr := config.AddNamespace(opt.Namespace, util.NowUTC())
			if delErr != nil || addErr != nil {
				return util.NewInternalError("Failed to overwrite namespace " + opt.Namespace)
			}
		}
	} else {
		// Create namespace
		if err := config.AddNamespace(opt.Namespace, util.NowUTC()); err != nil {
			return err
		}
	}
	// Flush at the end
	defer config.Flush()

	if opt.InputFile != "" {
		return executeWithYAML(opt.InputFile, opt.Namespace)
	}
	return manualExecute(opt)
}

func manualExecute(opt *Options) (err error) {
	if err := hasAllFlags(opt); err != nil {
		return err
	}

	// K8s or Remote
	var exe execute.Executor
	if opt.KubeConfig != "" {
		exe, err = connectk8scontrolplane.NewManualExecutor(opt.Namespace, opt.ControllerEndpoint, opt.KubeConfig, opt.IofogUserEmail, opt.IofogUserPass)
		if err != nil {
			return err
		}
	} else {
		exe, err = connectremotecontrolplane.NewManualExecutor(opt.Namespace, opt.ControllerName, opt.ControllerEndpoint, opt.IofogUserEmail, opt.IofogUserPass)
		if err != nil {
			return err
		}
	}

	// Execute
	if err := exe.Execute(); err != nil {
		return err
	}
	return nil
}

func executeWithYAML(yamlFile, namespace string) error {
	executorsMap, err := execute.GetExecutorsFromYAML(yamlFile, namespace, kindHandlers)
	if err != nil {
		return err
	}

	for idx := range kindOrder {
		if errs := execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("connect %s", kindOrder[idx])); len(errs) > 0 {
			return execute.CoalesceErrors(errs)
		}
	}

	return nil
}

func hasAllFlags(opt *Options) error {
	if opt.IofogUserEmail == "" || opt.IofogUserPass == "" {
		return util.NewInputError("Must provide ioFog User and Password flags")
	}
	if opt.KubeConfig == "" {
		if opt.ControllerName == "" || opt.ControllerEndpoint == "" {
			return util.NewInputError("Must provide Controller Name and Endpoint flags for Remote Control Plane")
		}
	} else {
		if opt.ControllerName != "" || opt.ControllerEndpoint != "" {
			return util.NewInputError("Cannot specify Controller Name and Endpoint for Kubernetes Control Plane")
		}
	}
	return nil
}

func generateConnectionString(namespace string) error {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return util.NewInputError("Cannot generate Connection String for non-existent Namespace")
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	if _, ok := controlPlane.(*rsc.LocalControlPlane); ok {
		return util.NewInputError("Cannot generate Connection String for Local Control Plane")
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return util.NewError("Could not get Control Plane endpoint")
	}
	msg := fmt.Sprintf("iofogctl connect --ecn-addr %s --name remote --email %s --pass %s --b64", endpoint, controlPlane.GetUser().Email, controlPlane.GetUser().Password)
	fmt.Println(msg)
	return nil
}
