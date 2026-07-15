package cmd

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/rollback"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newRollbackCommand() *cobra.Command {
	var opt rollback.Options

	cmd := &cobra.Command{
		Use:   "rollback RESOURCE NAME",
		Short: "Rollback ioFog resources",
		Long:  `Rollback ioFog resources to latest versions available.`,
		Example: ex(`%[1]s rollback agent NAME
%[1]s rollback agent NAME --semver v1.0.0`),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opt.ResourceType = args[0]
			opt.Name = args[1]

			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			opt.Semver, err = cmd.Flags().GetString("semver")
			util.Check(err)

			exe, err := rollback.NewExecutor(opt)
			util.Check(err)

			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully scheduled rollback for %s %s", strings.Title(opt.ResourceType), opt.Name)
			if opt.Semver != "" {
				msg += fmt.Sprintf(" to %s", opt.Semver)
			}
			util.PrintSuccess(msg)
		},
	}

	cmd.Flags().String("semver", "", "Target fog node version (semver.org; optional leading v)")

	return cmd
}
