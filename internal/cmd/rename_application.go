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
	rename "github.com/eclipse-iofog/iofogctl/internal/rename/application"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRenameApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application NAME NEW_NAME",
		Short:   "Rename a Application in your ECN to another name",
		Long:    `Rename a Application in your ECN to another name`,
		Example: `iofogctl rename application NAME NEW_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and the new name of the application
			name := args[0]
			newName := args[1]

			// Get an executor for the command
			err := rename.Execute("", name, newName)
			util.Check(err)

			util.PrintSuccess("Successfully renamed application " + name + " to " + newName)
		},
	}

	return cmd
}