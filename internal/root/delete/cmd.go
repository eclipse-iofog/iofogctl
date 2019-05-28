package delete

import (
	"github.com/eclipse-iofog/cli/internal/root/delete/agent"
	"github.com/eclipse-iofog/cli/internal/root/delete/controller"
	"github.com/eclipse-iofog/cli/internal/root/delete/microservice"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing ioFog resources",
		Long:  `Delete existing ioFog resources`,
	}

	cmd.AddCommand(deletecontroller.NewCommand())
	cmd.AddCommand(deleteagent.NewCommand())
	cmd.AddCommand(deletemicroservice.NewCommand())
	return cmd
}
