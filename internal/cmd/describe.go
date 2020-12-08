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
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	// Values accepted in resource type argument
	filename := ""
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Get detailed information of an existing resources",
		Long: `Get detailed information of an existing resources.
 
Most resources require a working Controller in the Namespace in order to be described.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDescribeControlPlaneCommand(),
		newDescribeControllerCommand(),
		newDescribeNamespaceCommand(),
		newDescribeAgentCommand(),
		newDescribeRegistryCommand(),
		newDescribeAgentConfigCommand(),
		newDescribeMicroserviceCommand(),
		newDescribeApplicationCommand(),
		newDescribeApplicationTemplateCommand(),
		newDescribeVolumeCommand(),
		newDescribeRouteCommand(),
		newDescribeEdgeResourceCommand(),
	)

	// Register Flags
	cmd.Flags().StringVarP(&filename, "output-file", "o", "", "YAML output file")

	return cmd
}
