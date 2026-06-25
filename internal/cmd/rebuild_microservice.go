package cmd

import (
	rebuildmicroservice "github.com/eclipse-iofog/iofogctl/internal/rebuild/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRebuildMicroserviceCommand() *cobra.Command {
	opt := rebuildmicroservice.Options{}
	cmd := &cobra.Command{
		Use:     "microservice AppNAME/MsvcNAME",
		Short:   "Rebuilds a microservice",
		Long:    "Rebuilds a microservice",
		Example: ex(`%[1]s rebuild microservice AppNAME/MsvcNAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify an microservice to rebuild"))
			}
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := rebuildmicroservice.NewExecutor(opt)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully rebuild Microservice " + opt.Name)
		},
	}
	return cmd
}
