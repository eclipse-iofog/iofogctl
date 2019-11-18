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
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/configure"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newConfigureCommand() *cobra.Command {
	// Values accepted in resource type argument
	var validResources = []string{"controller", "connector", "agent", "all", "agents", "controllers", "connectors", "default-namespace"}
	// Instantiate options
	var opt configure.Options

	cmd := &cobra.Command{
		Use:   "configure resource NAME",
		Short: "Configure iofogctl or SSH details an existing resource",
		Long: `Configure iofogctl or SSH details for an existing resource

Note that you cannot (and shouldn't need to) configure the host value of Agents.`,
		Example: `iofogctl configure default-namespace NAME
iofogctl configure controller NAME --host HOST --user USER --key KEYFILE --port PORTNUM
iofogctl configure connector NAME --host HOST --user USER --key KEYFILE --port PORTNUM
iofogctl configure controller NAME --kube KUBECONFIG
iofogctl configure connector NAME --kube KUBECONFIG
iofogctl configure agent NAME --user USER --key KEYFILE --port PORTNUM

iofogctl configure all --user USER --key KEYFILE --port PORTNUM
iofogctl configure controllers --host HOST NAME --user USER --key KEYFILE --port PORTNUM
iofogctl configure connectors --host HOST --user USER --key KEYFILE --port PORTNUM
iofogctl configure agents --user USER --key KEYFILE --port PORTNUM
` + fmt.Sprintf("\nValid resources are: %s\n", strings.Join(validResources, ", ")),
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify a resource to configure"))
			}
			// Get resource type and name
			opt.ResourceType = args[0]
			if len(args) > 1 {
				opt.Name = args[1]
			} else {
				if opt.ResourceType != "all" && opt.ResourceType != "controllers" && opt.ResourceType != "connectors" && opt.ResourceType != "agents" {
					util.Check(util.NewInputError("Must specify resource name if not configuring a group of resources"))
				}
			}

			var err error

			// Get namespace option
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for configure command
			exe, err := configure.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVar(&opt.Host, "host", "", "Hostname of remote host")
	cmd.Flags().StringVar(&opt.User, "user", "", "Username of remote host")
	cmd.Flags().StringVar(&opt.KeyFile, "key", "", "Path to private SSH key")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube", "", "Path to Kubernetes configuration file")
	cmd.Flags().IntVar(&opt.Port, "port", 0, "Port number that iofogctl uses to SSH into remote hosts")

	return cmd
}
