package deploy

import (
	"github.com/eclipse-iofog/cli/internal/root/deploy/agent"
	"github.com/eclipse-iofog/cli/internal/root/deploy/controller"
	"github.com/eclipse-iofog/cli/internal/root/deploy/microservice"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy ioFog stack on existing infrastructure",
		Long:  `Deploy ioFog stack on existing infrastructure`,
	}

	cmd.AddCommand(deploycontroller.NewCommand())
	cmd.AddCommand(deployagent.NewCommand())
	cmd.AddCommand(deploymicroservice.NewCommand())
	return cmd
}
