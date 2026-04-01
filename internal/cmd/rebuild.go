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

func newRebuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuilds a microservice or system-microservice",
		Long:  "Rebuilds a microservice or system-microservice",
	}

	// Add subcommands
	cmd.AddCommand(
		newRebuildMicroserviceCommand(),
		newRebuildSystemMicroserviceCommand(),
	)
	return cmd
}
