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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {

	// Root command
	var cmd = &cobra.Command{
		Use:   "iofogctl",
		Short: "ioFog Unified Command Line Interface",
		Long:  "ioFog Unified Command Line Interface",
	}

	// Initialize config filename
	cobra.OnInitialize(initConfig)

	// Global flags
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "CLI configuration file (default is ~/"+config.DefaultFilename+")")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(
		newConnectCommand(),
		newDisconnectCommand(),
		newDeployCommand(),
		newDeleteCommand(),
		newCreateCommand(),
		newGetCommand(),
		newDescribeCommand(),
		newLogsCommand(),
		newLegacyCommand(),
		newVersionCommand(),
	)

	return cmd
}

// Config file set by --config persistent flag
var configFilename string

// Callback for cobra on initialization
func initConfig() {
	config.Init(configFilename)
}
