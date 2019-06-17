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
	"sync"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Filename  string
	Namespace string
}

type input struct {
	Controllers   []config.Controller   `mapstructure:"controllers"`
	Agents        []config.Agent        `mapstructure:"agents"`
	Microservices []config.Microservice `mapstructure:"microservices"`
}

func Execute(opt *Options) error {
	// Check filename option
	if opt.Filename == "" {
		return util.NewInputError("Must specify resource definition filename")
	}

	// Check namespace option
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// Unmarshall the input file
	var in input
	err = util.UnmarshalYAML(opt.Filename, &in)
	if err != nil {
		return err
	}

	// If no controller is provided, one must already exist
	if len(in.Controllers) == 0 {
		if len(ns.Controllers) == 0 {
			return util.NewInputError("If you are not deploying a new controller, one must exist in the specified namespace")
		}
	}

	// Only support single controller
	if len(in.Controllers) > 1 {
		return util.NewInputError("Only single controller deployments are supported")
	}

	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup

	// Deploy controllers
	for _, ctrl := range in.Controllers {
		ctrlOpt := &deploycontroller.Options{
			Namespace:        opt.Namespace,
			Name:             ctrl.Name,
			User:             ctrl.User,
			Host:             ctrl.Host,
			Local:            util.IsLocalHost(ctrl.Host),
			KubeConfig:       ctrl.KubeConfig,
			KubeControllerIP: ctrl.KubeControllerIP,
			Images:           ctrl.Images,
		}
		exe, err := deploycontroller.NewExecutor(ctrlOpt)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := exe.Execute()
			util.Check(err)
		}()
	}
	wg.Wait()

	// Deploy agents
	for _, agent := range in.Agents {
		if agent.Port == 0 {
			agent.Port = 22
		}
		agentOpt := &deployagent.Options{
			Namespace: opt.Namespace,
			Name:      agent.Name,
			User:      agent.User,
			Host:      agent.Host,
			Port:      agent.Port,
			KeyFile:   agent.KeyFile,
			Local:     util.IsLocalHost(agent.Host),
			Image:     agent.Image,
		}
		exe, err := deployagent.NewExecutor(agentOpt)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := exe.Execute()
			util.Check(err)
		}()
	}
	wg.Wait()

	// TODO: Deploy microservices

	return config.Flush()
}
