package main

import (
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	// Root command
	var cmd = &cobra.Command{
		Use:   "iofogctl",
		Short: "ioFogctl Unified Command Line Interface",
		Long:  "ioFogctl Unified Command Line Interface",
	}

	// Initialize config filename
	cobra.OnInitialize(initConfig)

	// Global flags
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "CLI configuration file (default is ~/"+config.DefaultFilename+")")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(
		newGetCommand(),
		newDeployCommand(),
		newDeleteCommand(),
		newDescribeCommand(),
		newLogsCommand(),
	)

	return cmd
}

// Config file set by --config persistent flag
var configFilename string

// Callback for cobra on initialization
func initConfig() {
	config.SetFile(configFilename)
}
