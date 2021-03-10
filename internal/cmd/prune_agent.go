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
	prune "github.com/eclipse-iofog/iofogctl/v3/internal/prune/agent"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newPruneAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Remove all dangling images from Agent",
		Long:    `Remove all the images which are not used by existing containers on the specified Agent`,
		Example: `iofogctl prune agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// Run the command
			exe := prune.NewExecutor(namespace, name, useDetached)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully pruned " + name)
		},
	}

	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
