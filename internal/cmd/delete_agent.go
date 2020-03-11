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
	delete "github.com/eclipse-iofog/iofogctl/v2/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteAgentCommand() *cobra.Command {
	var soft bool
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Delete an Agent",
		Long: `Delete an Agent.

The Agent will be deleted from the Controller within the namespace.
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
			exe, _ := delete.NewExecutor(namespace, name, useDetached, soft)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted " + namespace + "/" + name)
		},
	}

	cmd.Flags().BoolVar(&soft, "soft", false, "Don't delete iofog-agent from remote host")

	return cmd
}
