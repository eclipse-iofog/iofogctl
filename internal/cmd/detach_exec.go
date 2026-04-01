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
	"fmt"

	detachagent "github.com/eclipse-iofog/iofogctl/internal/detach/exec/agent"
	detach "github.com/eclipse-iofog/iofogctl/internal/detach/exec/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func NewDetachExecMicroserviceCommand() *cobra.Command {
	opt := detach.Options{}
	cmd := &cobra.Command{
		Use:     "microservice NAME",
		Short:   "Detach an Exec Session to a Microservice",
		Long:    `Detach an Exec Session to an existing Microservice.`,
		Example: `iofogctl detach exec microservice AppName/MicroserviceName`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully detached Exec Session from Microservice %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newDetachExecAgentCommand() *cobra.Command {
	opt := detachagent.Options{}
	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Detach an Exec Session from an Agent",
		Long:    `Detach an Exec Session from an existing Agent.`,
		Example: `iofogctl detach exec agent AgentName`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := detachagent.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully detached Exec Session from Agent %s", opt.Name)
			util.PrintSuccess(msg)
		},
	}

	return cmd
}

func newDetachExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec",
		Short:   "Detach an Exec Session to a resource",
		Long:    `Detach an Exec Session to a Microservice or Agent.`,
		Example: `iofogctl detach exec microservice AppName/MicroserviceName`,
	}

	// Add subcommands
	cmd.AddCommand(
		NewDetachExecMicroserviceCommand(),
		newDetachExecAgentCommand(),
	)

	return cmd
}
