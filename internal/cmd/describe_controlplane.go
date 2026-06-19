package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeControlPlaneCommand() *cobra.Command {
	opt := describe.Options{
		Resource: "controlplane",
	}

	cmd := &cobra.Command{
		Use:     "controlplane",
		Short:   "Get detailed information about a Control Plane",
		Long:    `Get detailed information about the Control Plane in a single Namespace.`,
		Example: `iofogctl describe controlplane`,
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for describe command
			exe, err := describe.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&opt.Filename, "output-file", "o", "", "YAML output file")

	return cmd
}
