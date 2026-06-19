package cmd

import (
	startmicroservice "github.com/eclipse-iofog/iofogctl/internal/start/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newStartMicroserviceCommand() *cobra.Command {
	opt := startmicroservice.Options{}
	cmd := &cobra.Command{
		Use:     "microservice AppNAME/MsvcNAME",
		Short:   "Starts an microservice",
		Long:    "Starts an microservice",
		Example: `iofogctl start microservice AppNAME/MsvcNAME`,
		Args:    cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify an microservice to start"))
			}
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := startmicroservice.NewExecutor(opt)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully started Microservice " + opt.Name)
		},
	}
	return cmd
}
