/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
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
	deleterolebinding "github.com/eclipse-iofog/iofogctl/internal/delete/rolebinding"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteRoleBindingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rolebinding NAME",
		Short:   "Delete a RoleBinding",
		Long:    `Delete a RoleBinding from the Controller.`,
		Example: `iofogctl delete rolebinding NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deleterolebinding.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted rolebinding " + name)
		},
	}

	return cmd
}
