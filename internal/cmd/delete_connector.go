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
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteConnectorCommand() *cobra.Command {
	var soft bool
	cmd := &cobra.Command{
		Use:     "connector NAME",
		Short:   "Delete a Connector",
		Long:    `Delete a Connector.`,
		Example: `iofogctl delete connector NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of connector
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// Get an executor for the command
			exe, _ := delete.NewExecutor(namespace, name, useDetached, soft)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted " + namespace + "/" + name)
		},
	}

	cmd.Flags().BoolVar(&soft, "soft", false, "Don't delete iofog-connector from remote host")
	return cmd
}
