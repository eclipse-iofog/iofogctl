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

package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}
	filename := ""

	// Instantiate command
	cmd := &cobra.Command{
		Use: "deploy",
		Example: `deploy -f platform.yaml
deploy [command]`,
		Short: "Deploy ioFog platform or components on existing infrastructure",
		Long: `Deploy ioFog platform or individual components on existing infrastructure.

A YAML resource definition file can be use in lieu of the subcommands to deploy Controllers, Agents, and Microservices.

The YAML resource definition file should look like this (two Controllers specified for example only):
controllers:
- name: k8s
  kubeconfig: ~/.kube/conf
- name: vanilla 
  user: serge
  host: 35.239.157.151
  keyfile: ~/.ssh/id_rsa
agents:
- name: agent1
  user: serge
  host: 35.239.157.151
  keyfile: ~/.ssh/id_rsa
- name: agent2
  user: serge
  host: 35.232.114.32
  keyfile: ~/.ssh/id_rsa
microservices: []
`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			// Unmarshall the input file
			err = util.UnmarshalYAML(filename, &opt)
			util.Check(err)

			// Pre-process inputs
			for idx := range opt.Controllers {
				ctrl := &opt.Controllers[idx]
				// Fix SSH port
				if ctrl.Port == 0 {
					ctrl.Port = 22
				}
				// Format file paths
				ctrl.KeyFile, err = util.FormatPath(ctrl.KeyFile)
				util.Check(err)
				ctrl.KubeConfig, err = util.FormatPath(ctrl.KubeConfig)
				util.Check(err)
			}
			for idx := range opt.Agents {
				agent := &opt.Agents[idx]
				// Fix SSH port
				if agent.Port == 0 {
					agent.Port = 22
				}
				// Format file paths
				agent.KeyFile, err = util.FormatPath(agent.KeyFile)
				util.Check(err)
			}

			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			err = deploy.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed resources to namespace " + opt.Namespace)
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newDeployControllerCommand(),
		newDeployAgentCommand(),
		newDeployMicroserviceCommand(),
		newDeployApplicationCommand(),
	)

	// Register flags
	cmd.Flags().StringVarP(&filename, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
