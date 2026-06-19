package cmd

import (
	rebuildsystemmicroservice "github.com/eclipse-iofog/iofogctl/internal/rebuild/systemmicroservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRebuildSystemMicroserviceCommand() *cobra.Command {
	opt := rebuildsystemmicroservice.Options{}
	cmd := &cobra.Command{
		Use:     "system-microservice AppNAME/MsvcNAME",
		Short:   "Rebuilds a system microservice",
		Long:    "Rebuilds a system microservice",
		Example: `iofogctl rebuild system-microservice AppNAME/MsvcNAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify an microservice to rebuild"))
			}
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := rebuildsystemmicroservice.NewExecutor(opt)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully rebuild Microservice " + opt.Name)
		},
	}
	return cmd
}
