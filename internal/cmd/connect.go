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
	"github.com/eclipse-iofog/iofogctl/internal/connect"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newConnectCommand() *cobra.Command {
	//Instantiate options
	opt := connect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to an existing ioFog cluster",
		Long: `Connect to an existing ioFog cluster.

This command must be executed within an empty or non-existent namespace.
All resources provisioned with the corresponding Controller will become visible under the namespace.
All ssh access will be configured as provided in the config file.
See iofog.org for the YAML format.`,
		Example: `iofogctl connect -f platform.yaml
iofogctl connect --kube FILE --name NAME --email EMAIL --pass PASSWORD
iofogctl connect --ecn-addr ENDPOINT --name NAME --email EMAIL --pass PASSWORD`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			// Execute command
			err = connect.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully connected resources to namespace " + opt.Namespace)
		},
	}
	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")
	cmd.Flags().StringVar(&opt.ControllerName, "name", "", "Name you would like to assign to Controller")
	cmd.Flags().StringVar(&opt.ControllerEndpoint, "ecn-addr", "", "URL of Edge Compute Network to connect to")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube", "", "Kubernetes config file. Typically ~/.kube/config")
	cmd.Flags().StringVar(&opt.IofogUserEmail, "email", "", "ioFog user email address")
	cmd.Flags().StringVar(&opt.IofogUserPass, "pass", "", "ioFog user password")
	cmd.Flags().BoolVar(&opt.OverwriteNamespace, "force", false, "Overwrite existing namespace")

	return cmd
}
