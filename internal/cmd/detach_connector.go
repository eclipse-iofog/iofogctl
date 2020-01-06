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
	detach "github.com/eclipse-iofog/iofogctl/internal/detach/connector"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachConnectorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "connector NAME",
		Short:   "Detaches a Connector from its ECN",
		Long:    `Detaches a Connector from its ECN.`,
		Example: `iofogctl detach connector NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of connector
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			exe, _ := detach.NewExecutor(namespace, name)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully detached connector " + name)
		},
	}

	return cmd
}
