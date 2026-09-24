package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeModelCommand() *cobra.Command {
	opt := describe.Options{
		Resource: "model",
	}

	cmd := &cobra.Command{
		Use:     "model NAME",
		Short:   "Get detailed information about a Model",
		Long:    `Get detailed information about a Model.`,
		Example: ex(`%[1]s describe model NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Name = args[0]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := describe.NewExecutor(&opt)
			util.Check(err)

			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&opt.Filename, "output-file", "o", "", "YAML output file")

	return cmd
}
