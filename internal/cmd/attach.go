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

func newAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attach",
		Example: `attach`,
		Short:   "Attach an existing ioFog resource to an ECN",
		Long:    `Attach an existing ioFog resource to an ECN.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newAttachConnectorCommand(),
		newAttachAgentCommand(),
	)

	return cmd
}
