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
	detach "github.com/eclipse-iofog/iofogctl/v2/internal/detach/edgeresource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newDetachEdgeResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edge-resource NAME/VERSION AGENT_NAME",
		Short:   "Detaches an Edge Resource from an Agent",
		Long:    `Detaches an Edge Resource from an Agent.`,
		Example: `iofogctl detach edge-resource NAME/VERSION AGENT_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of edge resource
			nameVersion := args[0]
			agent := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detach.NewExecutor(namespace, nameVersion, agent)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully detached " + nameVersion)
		},
	}

	return cmd
}
