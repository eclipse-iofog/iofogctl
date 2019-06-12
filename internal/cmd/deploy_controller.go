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
	deploy "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployControllerCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "controller NAME",
		Short: "Deploy a Controller",
		Long: `Deploy a Controller.

On a Kubernetes deployment, this will install all resources under the iofog namespace.`,
		Example: `iofogctl deploy controller NAME --local 
 iofogctl deploy controller NAME --user root --host 32.23.134.3 --key_file ~/.ssh/id_ecdsa
iofogctl deploy controller NAME --kube-config ~/.kube/conf`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Get name and namespace of controller
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for command
			ctrl, err := deploy.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = ctrl.Execute()
			util.Check(err)
		},
	}

	// Set up options
	cmd.Flags().StringVarP(&opt.User, "user", "u", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.Host, "host", "o", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().StringVarP(&opt.KeyFile, "key-file", "k", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir")
	cmd.Flags().StringVarP(&opt.KubeConfig, "kube-config", "q", "", "Filename of Kubernetes cluster config file")
	cmd.Flags().StringVar(&opt.ImagesFile, "images", "", "Filename of YAML containing list of ioFog service images to be deployed on K8s cluster")
	cmd.Flags().BoolVarP(&opt.Local, "local", "l", false, "Configure for local deployment")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	return cmd
}
