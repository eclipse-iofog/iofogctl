package cmd

import (
	startapplication "github.com/eclipse-iofog/iofogctl/internal/start/application"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newStartApplicationCommand() *cobra.Command {
	opt := startapplication.Options{}
	cmd := &cobra.Command{
		Use:     "application NAME",
		Short:   "Starts an application",
		Long:    "Starts an application",
		Example: ex(`%[1]s start application NAME`),
		Args:    cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify an application to start"))
			}
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := startapplication.NewExecutor(opt)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully started Application " + opt.Name)
		},
	}
	return cmd
}
