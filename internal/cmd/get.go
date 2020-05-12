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
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/get"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	validResources := []string{"all", "namespaces", "controllers", "agents", "applications", "microservices", "catalog", "registries", "volumes", "routes"}
	cmd := &cobra.Command{
		Use:   "get RESOURCE",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources.

Resources like Agents will require a working Controller in the namespace to display all information.`,
		Example: `iofogctl get all
iofogctl get namespaces
iofogctl get controllers
iofogctl get agents
iofogctl get applications
iofogctl get microservices
iofogctl get catalog
iofogctl get registries
iofogctl get volumes
iofogctl get routes` + fmt.Sprintf("\n\nValid resources are: %s\n", strings.Join(validResources, ", ")),
		ValidArgs: validResources,
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			showDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			if showDetached && namespace != config.GetDefaultNamespaceName() {
				util.PrintNotify("You are requesting detached resources, namespace will be ignored.")
			}

			// Get executor for get command
			exe, err := get.NewExecutor(resource, namespace, showDetached)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
