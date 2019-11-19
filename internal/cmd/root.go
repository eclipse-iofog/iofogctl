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
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

const TitleHeader = "     _       ____                 __  __    \n" +
	"    (_)___  / __/___  ____  _____/ /_/ / 	 \n" +
	"   / / __ \\/ /_/ __ \\/ __ `/ ___/ __/ /   \n" +
	"  / / /_/ / __/ /_/ / /_/ / /__/ /_/ /   	 \n" +
	" /_/\\____/_/  \\____/\\__, /\\___/\\__/_/  \n" +
	"                   /____/                   \n"

const TitleMessage = "Welcome to the cool new iofogctl Cli!\n" +
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
	cmd.PersistentFlags().BoolVar(&httpVerbose, "http-verbose", false, "Toggle for displaying verbose output of API client")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(
		newConnectCommand(),
		newConfigureCommand(),
		newDisconnectCommand(),
		newDeployCommand(),
		newDeleteCommand(),
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
	)

	return cmd
}

// Namespace set by -n / --namespace
var namespace string

// Toggle set by --verbose persistent flag
var verbose bool

// Toggle set by --http-verbose persistent flag
var httpVerbose bool

// Callback for cobra on initialization
func initialize() {
	config.Init(namespace, "")
	client.SetVerbosity(httpVerbose)
	install.SetVerbosity(verbose)
	util.SpinEnable(!verbose && !httpVerbose)
}
