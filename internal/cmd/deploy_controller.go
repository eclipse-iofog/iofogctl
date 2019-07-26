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
	"fmt"
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
iofogctl deploy controller NAME --user root --host 32.23.134.3 --key_file ~/.ssh/id_rsa
iofogctl deploy controller NAME --kube-config ~/.kube/conf`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Get name and namespace of controller
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Format any file paths
			opt.KubeConfig, err = util.FormatPath(opt.KubeConfig)
			util.Check(err)
			opt.KeyFile, err = util.FormatPath(opt.KeyFile)
			util.Check(err)

			// Get executor for command
			ctrl, err := deploy.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = ctrl.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deployed " + opt.Namespace + "/" + opt.Name)
		},
	}

	// Set up options
	cmd.Flags().StringVar(&opt.User, "user", "", "Username of host the Controller is being deployed on")
	cmd.Flags().StringVar(&opt.Host, "host", "", "IP or hostname of host the Controller is being deployed on")
	cmd.Flags().IntVar(&opt.Port, "port", 22, "Port to use for SSH connection")
	cmd.Flags().StringVar(&opt.KeyFile, "key-file", "", "Filename of SSH private key used to access host. Corresponding *.pub must be in same dir. Must be RSA key.")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube-config", "", "Filename of Kubernetes cluster config file")
	cmd.Flags().StringVar(&opt.KubeControllerIP, "kube-controller-ip", "", "Static IP to assign to Kubernetes LoadBalancer")
	cmd.Flags().StringVar(&opt.ImagesFile, "images", "", "Filename of YAML containing list of ioFog service images to be deployed on K8s cluster")
	cmd.Flags().BoolVarP(&opt.Local, "local", "l", false, "Configure for local deployment")
	cmd.Flags().Lookup("local").NoOptDefVal = "true"

	cmd.SetHelpFunc(func(*cobra.Command, []string) {
		fmt.Print(helpMsg)
	})
	return cmd
}

const helpMsg = `Deploy a Controller either locally, remotely, or on a Kubernetes cluster.

On a Kubernetes deployment, this will install all resources under the namespace specified by the namespace flag of this command.

Usage:
  iofogctl deploy controller NAME [flags]

Examples:
iofogctl deploy controller NAME --local 
iofogctl deploy controller NAME --user root --host 32.23.134.3 --key_file ~/.ssh/id_rsa
iofogctl deploy controller NAME --kube-config ~/.kube/conf

Remote Deploy Flags:
      --host string                 IP or hostname of host the Controller is being deployed on
      --user string                 Username of host the Controller is being deployed on
      --key-file string             Filename of SSH private key used to access host. Corresponding *.pub must be in same dir. Must be RSA key.
      --port int                    Port to use for SSH connection (default 22)

Kubernetes Deploy Flags:
      --kube-config string          Filename of Kubernetes cluster config file
      --images string               Filename of YAML containing list of ioFog service images to be deployed on K8s cluster
      --kube-controller-ip string   Static IP to assign to Kubernetes LoadBalancer

Local Deploy Flags:
  -l, --local                       Configure for local deployment

Common Flags:
  -h, --help                        help for controller

Global Flags:
      --config string      CLI configuration file (default is ~/.iofog/config.yaml)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -q, --quiet              Toggle for displaying verbose output
`
