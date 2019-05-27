package deploy

import (
	"github.com/spf13/cobra"
	"github.com/eclipse-iofog/cli/pkg/root/deploy/controller"
	"github.com/eclipse-iofog/cli/pkg/root/deploy/agent"
	"github.com/eclipse-iofog/cli/pkg/root/deploy/microservice"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy ioFog stack on existing infrastructure",
		Long: `Deploy ioFog stack on existing infrastructure`,
	}

	cmd.AddCommand(deploycontroller.NewCommand())
	cmd.AddCommand(deployagent.NewCommand())
	cmd.AddCommand(deploymicroservice.NewCommand())
	return cmd
}