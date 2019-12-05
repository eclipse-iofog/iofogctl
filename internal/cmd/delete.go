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
	"github.com/eclipse-iofog/iofogctl/internal/delete"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	// Instantiate options
	opt := &delete.Options{}

	cmd := &cobra.Command{
		Use:     "delete",
		Example: `deploy -f platform.yaml`,
		Short:   "Delete an existing ioFog resource",
		Long:    `Delete an existing ioFog resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			// Execute command
			err = delete.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deleted resources from namespace " + opt.Namespace)
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newDeleteNamespaceCommand(),
		newDeleteControllerCommand(),
		newDeleteConnectorCommand(),
		newDeleteAgentCommand(),
		newDeleteAllCommand(),
		newDeleteApplicationCommand(),
		newDeleteCatalogItemCommand(),
		newDeleteMicroserviceCommand(),
	)

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
