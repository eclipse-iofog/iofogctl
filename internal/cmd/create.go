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

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Long: `Create a resource.

Some resources (e.g. namespaces) are relevant to iofogctl alone. Other resources are created on the ioFog cluster itself.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newCreateNamespaceCommand(),
	)
	return cmd
}
