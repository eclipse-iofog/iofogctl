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
	"fmt"
	"sync"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace     string
	Controllers   []config.Controller   `mapstructure:"controllers"`
	Agents        []config.Agent        `mapstructure:"agents"`
	Microservices []config.Microservice `mapstructure:"microservices"`
}

type agentJobResult struct {
	agentConfig config.Agent
	err         error
}

func deployControllers(namespace string, controllers []config.Controller) (err error) {
	// Only support single controller
	if len(controllers) > 1 {
		return util.NewInputError("Only single controller deployments are supported")
	}

	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup

	// Deploy controllers
	for _, ctrl := range controllers {
		ctrlOpt := &deploycontroller.Options{
			Namespace:        namespace,
			Name:             ctrl.Name,
			User:             ctrl.User,
			Host:             ctrl.Host,
			Local:            util.IsLocalHost(ctrl.Host),
			KubeConfig:       ctrl.KubeConfig,
			KubeControllerIP: ctrl.KubeControllerIP,
			Images:           ctrl.Images,
			IofogUser:        ctrl.IofogUser,
			KeyFile:          ctrl.KeyFile,
			Port:             ctrl.Port,
		}

		var exe deploycontroller.Executor
		exe, err = deploycontroller.NewExecutor(ctrlOpt)
		if err != nil {
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := exe.Execute()
			util.Check(err)
		}()
	}
	wg.Wait()
	return
}

func deployAgents(namespace string, agents []config.Agent) error {
	// Instantiate wait group for parallel tasks
	var wg sync.WaitGroup
	localAgentCount := 0
	agentChan := make(chan agentJobResult, len(agents))
	for idx, agent := range agents {

		// Check local deploys
		local := false
		if util.IsLocalHost(agent.Host) {
			local = true
			localAgentCount++
			if localAgentCount > 1 {
				fmt.Printf("Agent [%v] not deployed, you can only run one local agent.\n", agent.Name)
				continue
			}
		}
		agentOpt := &deployagent.Options{
			Namespace: namespace,
			Name:      agent.Name,
			User:      agent.User,
			Host:      agent.Host,
			Port:      agent.Port,
			KeyFile:   agent.KeyFile,
			Local:     local,
			Image:     agent.Image,
		}

		var exe deployagent.Executor
		exe, err := deployagent.NewExecutor(agentOpt)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			err := exe.Execute()
			agentChan <- agentJobResult{
				err: err,
			}
		}(idx, agent.Name)
	}
	wg.Wait()
	close(agentChan)

	// Output any errors
	failed := false
	for agentJobResult := range agentChan {
		if agentJobResult.err != nil {
			failed = true
			util.PrintNotify(agentJobResult.err.Error())
		}
	}

	if failed {
		return util.NewError("Failed to deploy one or more resources")
	}

	return nil
}

func Execute(opt *Options) error {
	// Check namespace option
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// If there are no resources return error
	if len(opt.Controllers) == 0 && len(opt.Agents) == 0 {
		return util.NewInputError("No resources specified to deploy in the YAML file")
	}

	// If no controller is provided, one must already exist
	if len(opt.Controllers) == 0 {
		if len(ns.Controllers) == 0 {
			return util.NewInputError("If you are not deploying a new controller, one must exist in the specified namespace")
		}
	}

	// Deploy Controllers
	if err = deployControllers(opt.Namespace, opt.Controllers); err != nil {
		return err
	}

	// Deploy Agents
	if err = deployAgents(opt.Namespace, opt.Agents); err != nil {
		return err
	}

	return nil
}
