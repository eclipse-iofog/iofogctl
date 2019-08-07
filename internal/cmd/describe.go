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

	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	validResources := make([]string, len(resources))
	i := 0
	for resource := range resources {
		validResources[i] = resource
		i++
	}
	filename := ""
	cmd := &cobra.Command{
		Use:   "describe resource NAME",
		Short: "Get detailed information of existing resources",
		Long: `Get detailed information of existing resources.

Resources such as Agents require a working Controller in the namespace in order to be described.`,
		Example: `iofogctl describe controller NAME
iofogctl describe agent NAME
iofogctl describe microservice NAME` + fmt.Sprintf("\n\nValid resources are: %s\n", strings.Join(validResources, ", ")),
		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			resource := args[0]
			name := args[1]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Validate first argument
			if _, exists := resources[resource]; !exists {
				util.Check(util.NewNotFoundError(resource))
			}

			// Get executor for describe command
			exe, err := describe.NewExecutor(resource, namespace, name, filename)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&filename, "output-file", "o", "", "YAML output file")

	return cmd
}
