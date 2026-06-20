package cmd

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/upgrade"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
	// Instantiate options
	var opt upgrade.Options

	cmd := &cobra.Command{
		Use:     "upgrade RESOURCE NAME",
		Short:   "Upgrade ioFog resources",
		Long:    `Upgrade ioFog resources to latest versions available.`,
		Example: `iofogctl upgrade agent NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			opt.ResourceType = args[0]
			opt.Name = args[1]

			var err error
			// Get namespace option
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for upgrade command
			exe, err := upgrade.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess(fmt.Sprintf("Successfully scheduled upgrade for %s %s", strings.Title(opt.ResourceType), opt.Name))
		},
	}

	return cmd
}
