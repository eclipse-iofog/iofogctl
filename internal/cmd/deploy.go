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
		Use:     "deploy",
		Example: `deploy -f platform.yaml`,
		Short:   "Deploy ioFog platform or components on existing infrastructure",
		Long: `Deploy ioFog platform or individual components on existing infrastructure.

The YAML resource specification file should look like this (two Controllers specified for example only):` + "\n```\n" +
			`controlplane:
	controllers:
	- name: k8s # Controller name
	  kubeconfig: ~/.kube/conf # Will deploy a controller in a kubernetes cluster
	- name: vanilla 
	  user: serge # SSH user
	  host: 35.239.157.151 # SSH Host - Will deploy a controller as a standalone binary
	  keyfile: ~/.ssh/id_rsa # SSH private key
	agents:
	- name: agent1 # Agent name
	  user: serge # SSH User
	  host: 35.239.157.151 # SSH host
	  keyfile: ~/.ssh/id_rsa # SSH private key
	- name: agent2
	  user: serge
	  host: 35.232.114.32
	  keyfile: ~/.ssh/id_rsa
	applications: [] # See iofogctl deploy application for an application yaml schema
	microservices: [] # See iofogctl deploy microservices
` + "\n```\n",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
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
		newDeployControlPlaneCommand(),
		newDeployAgentCommand(),
		newDeployApplicationCommand(),
	)

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
