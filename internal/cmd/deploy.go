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

	// Instantiate command
	cmd := &cobra.Command{
		Use: "deploy",
		Example: `deploy -f platform.yaml
deploy [command]`,
		Short: "Deploy ioFog platform or components on existing infrastructure",
		Long: `Deploy ioFog platform or individual components on existing infrastructure.

A YAML resource definition file can be use in lieu of the subcommands to deploy Controllers, Agents, and Microservices.

The YAML resource definition file should look like this:
controllers:
- name: sergek8s
  kubeconfig: ~/.kube/conf
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
			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

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
	)

	// Register flags
	cmd.Flags().StringVarP(&opt.Filename, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
