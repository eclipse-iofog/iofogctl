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

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Hidden:  true,
		Short:   "Updates a resource",
		Long:    "Updates a resource",
		Example: `iofogctl update`,
	}

	// Add subcommands
	cmd.AddCommand(
		newUpdateIofogctlCommand(),
	)
	return cmd
}
