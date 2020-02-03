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

func newDetachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "detach",
		Example: `detach`,
		Short:   "Detach an existing ioFog resource from its ECN",
		Long:    `Detach an existing ioFog resource from its ECN.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDetachAgentCommand(),
	)

	return cmd
}
