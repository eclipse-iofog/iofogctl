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
	delete "github.com/eclipse-iofog/iofogctl/v3/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteAgentCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Delete an Agent",
		Long: `Delete an Agent.

The Agent will be unprovisioned from the Controller within the namespace.

The Agent stack will be uninstalled from the host.

If you wish to not remove the Agent stack from the host, please use iofogctl detach agent`,
		Example: `iofogctl delete agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// Run the command
			exe, err := delete.NewExecutor(namespace, name, useDetached, force)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			printName := name
			if !useDetached {
				printName = namespace + "/" + name
			}
			util.PrintSuccess("Successfully deleted " + printName)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Remove even if there are still Microservices running on the Agent")
	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
