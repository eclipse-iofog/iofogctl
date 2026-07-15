package cmd

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/upgrade"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
	var opt upgrade.Options

	cmd := &cobra.Command{
		Use:   "upgrade RESOURCE NAME",
		Short: "Upgrade ioFog resources",
		Long:  `Upgrade ioFog resources to latest versions available.`,
		Example: ex(`%[1]s upgrade agent NAME
%[1]s upgrade agent NAME --semver v1.0.0`),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			opt.ResourceType = args[0]
			opt.Name = args[1]

			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			opt.Semver, err = cmd.Flags().GetString("semver")
			util.Check(err)

			exe, err := upgrade.NewExecutor(opt)
			util.Check(err)

			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully scheduled upgrade for %s %s", strings.Title(opt.ResourceType), opt.Name)
			if opt.Semver != "" {
				msg += fmt.Sprintf(" to %s", opt.Semver)
			}
			util.PrintSuccess(msg)
		},
	}

	cmd.Flags().String("semver", "", "Target fog node version (semver.org; optional leading v)")

	return cmd
}
