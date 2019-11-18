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

	"github.com/eclipse-iofog/iofogctl/internal/get"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	validResources := []string{"all", "namespaces", "controllers", "connectors", "agents", "applications", "microservices", "catalog"}
	cmd := &cobra.Command{
		Use:   "get RESOURCE",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources.

Resources like Agents will require a working Controller in the namespace to display all information.`,
		Example: `iofogctl get all
iofogctl get namespaces
iofogctl get controllers` + fmt.Sprintf("\n\nValid resources are: %s\n", strings.Join(validResources, ", ")),
		ValidArgs: validResources,
		Args:      cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]

			// Get executor for get command
			exe, err := get.NewExecutor(resource, "")
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
