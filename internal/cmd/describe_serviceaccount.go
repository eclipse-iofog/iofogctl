package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/describe"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeServiceAccountCommand() *cobra.Command {
	opt := describe.Options{
		Resource: "serviceaccount",
	}

	cmd := &cobra.Command{
		Use:     "serviceaccount APPLICATION_NAME/SERVICE_ACCOUNT_NAME",
		Short:   "Get detailed information about a ServiceAccount",
		Long:    `Get detailed information about a ServiceAccount. ServiceAccounts are application-scoped; use APPLICATION_NAME/SERVICE_ACCOUNT_NAME (e.g. myapp/my-sa).`,
		Example: ex(`%[1]s describe serviceaccount myapp/my-sa`),
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
