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
	Filename  string
	Namespace string
}

type input struct {
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
		}
		// Format file paths
		if ctrlOpt.KubeConfig, err = util.FormatPath(ctrlOpt.KubeConfig); err != nil {
			return 
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
		if agent.Port == 0 {
			agent.Port = 22
		}
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
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			agentConfig, err := deployagent.DeployAgent(agentOpt)
			agentChan <- agentJobResult{
				agentConfig: agentConfig,
				err:         err,
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
		} else {
			if err := config.UpdateAgent(namespace, agentJobResult.agentConfig); err != nil {
				util.PrintNotify("Failed to update config file but Agent " + agentJobResult.agentConfig.Name + " deployed successfully")
			}
		}
	}
	if err := config.Flush(); err != nil {
		util.PrintNotify("Failed to write to config file but resources were deployed")
	}

	if failed {
		return util.NewError("Failed to deploy one or more resources")
	}

	return nil
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

	// Deploy Controllers
	if err = deployControllers(opt.Namespace, in.Controllers); err != nil {
		return err
	}

	// Deploy Agents
	if err = deployAgents(opt.Namespace, in.Agents); err != nil {
		return err
	}

	return nil
}
