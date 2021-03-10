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
	rename "github.com/eclipse-iofog/iofogctl/v3/internal/rename/namespace"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newRenameNamespaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace NAME NEW_NAME",
		Short:   "Rename a Namespace",
		Long:    `Rename a Namespace`,
		Example: `iofogctl rename namespace NAME NEW_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and new name of the namespace
			name := args[0]
			newName := args[1]

			// Get an executor for the command
			err := rename.Execute(name, newName)
			util.Check(err)

			util.PrintSuccess(getRenameSuccessMessage("Namespace", name, newName))
		},
	}

	return cmd
}
