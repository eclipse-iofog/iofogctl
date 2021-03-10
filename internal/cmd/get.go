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
	"errors"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/get"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	validResources := []string{
		"all",
		"namespaces",
		"controllers",
		"agents",
		"edge-resources",
		"application-templates",
		"applications",
		"microservices",
		"catalog",
		"registries",
		"volumes",
		"routes",
	}
	cmd := &cobra.Command{
		Use:   "get RESOURCE",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources.

Resources like Agents will require a working Controller in the namespace to display all information.`,
		Example: `iofogctl get all
             namespaces
             controllers
             agents
             edge-resources
             application-templates
             applications
             microservices
             catalog
             registries
             volumes
             routes`,
		ValidArgs: validResources,
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			showDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// TODO: Break out resources as subcommands to avoid this kind of logic and improve --help accuracy
			if showDetached && resource != "agents" {
				err = errors.New("can only use --detached flag with Agents")
				util.Check(err)
			}

			if showDetached && namespace != config.GetDefaultNamespaceName() {
				util.PrintNotify("You are requesting detached resources, Namespace will be ignored.")
			}

			// Get executor for get command
			exe, err := get.NewExecutor(resource, namespace, showDetached)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}

	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
