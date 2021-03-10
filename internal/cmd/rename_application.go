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
	rename "github.com/eclipse-iofog/iofogctl/v3/internal/rename/application"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newRenameApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application NAME NEW_NAME",
		Short:   "Rename an Application",
		Long:    `Rename a Application`,
		Example: `iofogctl rename application NAME NEW_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and the new name of the application
			name := args[0]
			newName := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			err = rename.Execute(namespace, name, newName)
			util.Check(err)

			util.PrintSuccess(getRenameSuccessMessage("Application", name, newName))
		},
	}

	return cmd
}
