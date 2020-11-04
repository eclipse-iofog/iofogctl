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
	delete "github.com/eclipse-iofog/iofogctl/v2/internal/delete/edge-resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteEdgeResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge-resource NAME/VERSION",
		Short: "Delete an Edge Resource",
		Long: `Delete an Edge Resource.

Only the specified version will be deleted.
Agents that this Edge Resource are attached to will be notified of the deletion.`,
		Example: `iofogctl delete edge-resource NAME/VERSION`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of edge resource
			nameVersion := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := delete.NewExecutor(namespace, nameVersion)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted " + namespace + "/" + nameVersion)
		},
	}

	return cmd
}
