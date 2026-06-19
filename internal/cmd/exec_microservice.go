package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/exec"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newExecMicroserviceCommand() *cobra.Command {
	opt := exec.Options{
		Resource: "microservice",
	}

	cmd := &cobra.Command{
		Use:     "microservice AppName/MsvcName",
		Short:   "Connect to an Exec Session of a Microservice",
		Long:    `Connect to an Exec Session of a Microservice to interact with its container.`,
		Example: `iofogctl exec microservice AppName/MicroserviceName`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			var err error
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for exec command
			exe, err := exec.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
