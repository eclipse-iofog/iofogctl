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
	opt := &connect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "connect CONTROLLERNAME",
		Short: "Connect to an existing ioFog cluster",
		Long: `Connect to an existing ioFog cluster.

This command must be executed within an empty or non-existent namespace.
All resources provisioned with the corresponding Controller will become visible under the namespace.`,
		Example: `iofogctl connect CONTROLLERNAME --controller 123.321.123.22 --email EMAIL --pass PASSWORD
iofogctl connect CONTROLLERNAME --kube-config ~/.kube/conf --email EMAIL --pass PASSWORD`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource name
			opt.Name = args[0]

			// Get namespace option
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Format any file paths
			opt.KubeFile, err = util.FormatPath(opt.KubeFile)
			util.Check(err)

			// Get executor for get command
			exe, err := connect.NewExecutor(opt)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully connected to " + opt.Namespace + "/" + opt.Name)
		},
	}
	cmd.Flags().StringVarP(&opt.Endpoint, "controller", "c", "", "Host and (optionally) port of the Controller you are connecting to")
	cmd.Flags().StringVarP(&opt.KubeFile, "kube-config", "u", "", "Filename of Kubernetes cluster config file")
	cmd.Flags().StringVarP(&opt.Email, "email", "e", "", "Email address of user registered against Controller")
	cmd.Flags().StringVarP(&opt.Password, "pass", "p", "", "Password of user registered against Controller")

	return cmd
}
