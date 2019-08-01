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
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing ioFog resource",
		Long: `Delete an existing ioFog resource.

Deleting Agents or Controllers will result in the respective deployments being torn down.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeleteNamespaceCommand(),
		newDeleteControllerCommand(),
		newDeleteAgentCommand(),
		newDeleteMicroserviceCommand(),
		newDeleteAllCommand(),
		newDeleteApplicationCommand(),
	)
	return cmd
}
