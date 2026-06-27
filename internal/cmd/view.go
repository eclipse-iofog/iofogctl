package cmd

import (
	"fmt"
	"os"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Open EdgeOps Console",
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			ns, err := config.GetNamespace(namespace)
			util.Check(err)
			if len(ns.GetControllers()) == 0 {
				util.PrintError("You must deploy a Control Plane to a namespace to open the EdgeOps Console")
				os.Exit(1)
			}
			cp, err := ns.GetControlPlane()
			util.Check(err)
			consoleURL, err := resource.ResolveConsoleURL(cp)
			if err != nil {
				util.PrintError("Failed to resolve EdgeOps Console URL: " + err.Error())
				os.Exit(1)
			}
			if err := browser.OpenURL(consoleURL); err != nil {
				util.PrintInfo("To open the EdgeOps Console, go to:\n")
				util.PrintInfo(fmt.Sprintf("%s\n", consoleURL))
			}
		},
	}
	return cmd
}
