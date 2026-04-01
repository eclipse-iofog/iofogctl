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
	"github.com/spf13/cobra"
)

func newExecCommand() *cobra.Command {
	// Values accepted in resource type argument
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Connect to an Exec Session of a resource",
		Long:  `Connect to an Exec Session of a Microservice or Agent.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newExecMicroserviceCommand(),
		newExecAgentCommand(),
	)

	return cmd
}
