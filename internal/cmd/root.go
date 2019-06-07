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
		newGetCommand(),
		newDeployCommand(),
		newDeleteCommand(),
		newCreateCommand(),
		newDescribeCommand(),
		newLogsCommand(),
		newLegacyCommand(),
	)

	return cmd
}

// Config file set by --config persistent flag
var configFilename string

// Callback for cobra on initialization
func initConfig() {
	config.Init(configFilename)
}
