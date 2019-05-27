package root

import (
	"github.com/spf13/cobra"
	"github.com/eclipse-iofog/cli/pkg/root/get"
	"github.com/eclipse-iofog/cli/pkg/root/deploy"
	"github.com/eclipse-iofog/cli/pkg/root/delete"
	"github.com/eclipse-iofog/cli/pkg/root/describe"
	"github.com/eclipse-iofog/cli/pkg/root/logs"
	"github.com/eclipse-iofog/cli/pkg/config"
)

//NewCommand export
func NewCommand() *cobra.Command {
	// Root command
	var cmd = &cobra.Command{
		Use:   "iofog",
		Short: "ioFog Unified Command Line Interface",
		Long: "ioFog Unified Command Line Interface",
	}

	// Initialize config filename
	cobra.OnInitialize(initConfig)

	// Global flags
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "CLI configuration file (default is ~/" + config.DefaultFilename + ")")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(get.NewCommand())
	cmd.AddCommand(deploy.NewCommand())
	cmd.AddCommand(delete.NewCommand())
	cmd.AddCommand(describe.NewCommand())
	cmd.AddCommand(logs.NewCommand())

	return cmd
}

var configFilename string
func initConfig(){
	config.SetFile(configFilename)
}