package cmd

import (
	stopapplication "github.com/eclipse-iofog/iofogctl/internal/stop/application"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newStopApplicationCommand() *cobra.Command {
	opt := stopapplication.Options{}
	cmd := &cobra.Command{
		Use:     "application NAME",
		Short:   "Stop an application",
		Long:    "Stop an application",
		Example: ex(`%[1]s stop application NAME`),
		Args:    cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify an application to start"))
			}
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := stopapplication.NewExecutor(opt)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully stopped Application " + opt.Name)
		},
	}
	return cmd
}
