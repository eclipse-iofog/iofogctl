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
	move "github.com/eclipse-iofog/iofogctl/v2/internal/move/microservice"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newMoveMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microservice NAME AGENT_NAME",
		Short:   "Move a Microservice to another agent in the same ECN",
		Long:    `Move a Microservice to another agent in the same ECN`,
		Example: `iofogctl move microservice NAME AGENT_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace
			name := args[0]
			agent := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get an executor for the command
			err = move.Execute(namespace, name, agent)
			util.Check(err)

			util.PrintSuccess("Successfully moved microservice " + name + " to agent " + agent)
		},
	}

	return cmd
}
