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
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

const TitleHeader = "     _       ____                 __  __    \n" +
	"    (_)___  / __/___  ____  _____/ /_/ / 	 \n" +
	"   / / __ \\/ /_/ __ \\/ __ `/ ___/ __/ /   \n" +
	"  / / /_/ / __/ /_/ / /_/ / /__/ /_/ /   	 \n" +
	" /_/\\____/_/  \\____/\\__, /\\___/\\__/_/  \n" +
	"                   /____/                   \n"

const TitleMessage = "iofogctl is the CLI for ioFog. Think of it as a mix between terraform and kubectl.\n" +
	"\n" +
	"Use `iofogctl version` to display the current version.\n\n"

func printHeader() {
	util.PrintInfo(TitleHeader)
	util.PrintInfo("\n")
	util.PrintInfo(TitleMessage)
}

func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "iofogctl",
		//Short: "ioFog Unified Command Line Interface",
		PreRun: func(cmd *cobra.Command, args []string) {
			printHeader()
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.SetArgs([]string{"-h"})
			err := cmd.Execute()
			util.Check(err)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Initialize config filename
	cobra.OnInitialize(initialize)

	// Global flags
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Toggle for displaying verbose output of iofogctl")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Toggle for displaying verbose output of API clients (HTTP and SSH)")
	cmd.PersistentFlags().StringP("namespace", "n", config.GetDefaultNamespaceName(), "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(
		newConnectCommand(),
		newConfigureCommand(),
		newDisconnectCommand(),
		newDeployCommand(),
		newDeleteCommand(),
		newDetachCommand(),
		newAttachCommand(),
		newCreateCommand(),
		newGetCommand(),
		newDescribeCommand(),
		newLogsCommand(),
		newLegacyCommand(),
		newVersionCommand(),
		newBashCompleteCommand(cmd),
		newGenerateDocumentationCommand(cmd),
		newViewCommand(),
		newStartCommand(),
		newStopCommand(),
		newMoveCommand(),
		newRenameCommand(),
		newDockerPruneCommand(),
		newUpgradeCommand(),
		newRollbackCommand(),
	)

	return cmd
}

// Toggle set by --verbose persistent flag
var verbose bool

// Toggle set by --debug persistent flag
var debug bool

// Callback for cobra on initialization
func initialize() {
	client.SetGlobalRetries(client.Retries{
		Timeout: 20,
		CustomMessage: map[string]int{
			"timeout":                   20, // Linux
			"failed to respond":         20, // Windows
			"Bad Gateway":               20, // K8s
			"context deadline exceeded": 20,
		},
	})
	client.SetVerbosity(debug)
	install.SetVerbosity(verbose)
	util.SpinEnable(!verbose && !debug)
	util.SetDebug(debug)
}
