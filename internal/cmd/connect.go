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

package cmd

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/connect"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newConnectCommand() *cobra.Command {
	// Instantiate options
	opt := connect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to an existing Control Plane",
		Long: `Connect to an existing Control Plane.

This command must be executed within an empty or non-existent Namespace.
All resources provisioned with the corresponding Control Plane will become visible under the Namespace.
Visit iofog.org to view all YAML specifications usable with this command.`,
		Example: `iofogctl connect -f controlplane.yaml

iofogctl connect --email EMAIL --pass PASSWORD --kube     FILE 
                 --email EMAIL --pass PASSWORD --ecn-addr ENDPOINT --name NAME

iofogctl connect --generate`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			// Execute command
			err = connect.Execute(&opt)
			util.Check(err)

			if !opt.Generate {
				util.PrintSuccess("Successfully connected resources to namespace " + opt.Namespace)
			}
		},
	}
	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", pkg.flagDescYaml)
	cmd.Flags().StringVar(&opt.ControllerName, "name", "", "Name you would like to assign to Controller")
	cmd.Flags().StringVar(&opt.ControllerEndpoint, "ecn-addr", "", "URL of Edge Compute Network to connect to")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube", "", "Kubernetes config file. Typically ~/.kube/config")
	cmd.Flags().StringVar(&opt.IofogUserEmail, "email", "", "ioFog user email address")
	cmd.Flags().StringVar(&opt.IofogUserPass, "pass", "", "ioFog user password")
	cmd.Flags().BoolVar(&opt.OverwriteNamespace, "force", false, "Overwrite existing Namespace")
	cmd.Flags().BoolVar(&opt.Generate, "generate", false, "Generate a connection string that can be used to connect to this ECN")
	cmd.Flags().BoolVar(&opt.Base64Encoded, "b64", false, "Indicate whether input password (--pass) is base64 encoded or not")

	return cmd
}
