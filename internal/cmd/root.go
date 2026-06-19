package cmd

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

const iofogctlTitleHeader = "     _       ____                 __  __    \n" +
	"    (_)___  / __/___  ____  _____/ /_/ / 	 \n" +
	"   / / __ \\/ /_/ __ \\/ __ `/ ___/ __/ /   \n" +
	"  / / /_/ / __/ /_/ / /_/ / /__/ /_/ /   	 \n" +
	" /_/\\____/_/  \\____/\\__, /\\___/\\__/_/  \n" +
	"                   /____/                   \n"

const iofogctlTitleMessage = "iofogctl is the CLI for ioFog. Think of it as a mix between terraform and kubectl.\n" +
	"\n" +
	"Use `iofogctl version` to display the current version.\n\n"

const potctlTitleHeader = "\n" +
	"██████╗  ██████╗ ████████╗ ██████╗████████╗██╗     \n" +
	"██╔══██╗██╔═══██╗╚══██╔══╝██╔════╝╚══██╔══╝██║     \n" +
	"██████╔╝██║   ██║   ██║   ██║        ██║   ██║     \n" +
	"██╔═══╝ ██║   ██║   ██║   ██║        ██║   ██║     \n" +
	"██║     ╚██████╔╝   ██║   ╚██████╗   ██║   ███████╗\n" +
	"╚═╝      ╚═════╝    ╚═╝    ╚═════╝   ╚═╝   ╚══════╝\n"

const potctlTitleMessage = "potctl is the CLI for Datasance PoT. Think of it as a mix between terraform and kubectl.\n" +
	"\n" +
	"Use `potctl version` to display the current version.\n\n"

func printHeader() {
	if util.GetCliBinaryName() == "potctl" {
		util.PrintInfo(potctlTitleHeader)
		util.PrintInfo("\n")
		util.PrintInfo(potctlTitleMessage)
	} else {
		util.PrintInfo(iofogctlTitleHeader)
		util.PrintInfo("\n")
		util.PrintInfo(iofogctlTitleMessage)
	}
}

func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use: util.GetCliBinaryName(),
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
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Toggle for displaying verbose output of "+util.GetCliBinaryName())
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
		newRebuildCommand(),
		newRenameCommand(),
		newDockerPruneCommand(),
		newUpgradeCommand(),
		newRollbackCommand(),
		newExecCommand(),
		newNatsCommand(),
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
