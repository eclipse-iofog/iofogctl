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

func newDetachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "detach",
		Example: `detach`,
		Short:   "Detach one ioFog resource from another",
		Long:    `Detach one ioFog resource from another.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDetachAgentCommand(),
		newDetachEdgeResourceCommand(),
	)

	return cmd
}
