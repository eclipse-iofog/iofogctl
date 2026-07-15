package cmd

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get CLI application version",
		Run: func(cmd *cobra.Command, args []string) {
			ecnFlag, err := cmd.Flags().GetBool("ecn")
			util.Check(err)
			util.PrintInfo(fmt.Sprintf("%s - Copyright (C) 2026 Contributors\n", util.GetCliBinaryName()))
			_ = util.Print(util.GetVersion())
			if ecnFlag {
				fmt.Println("")
				fmt.Println("edgelet@" + util.GetEdgeletVersion())
				fmt.Println("")
				fmt.Println(util.GetControllerImage())
				fmt.Println(util.GetEdgeletImage())
				fmt.Println(util.GetOperatorImage())
				fmt.Println(util.GetRouterImage())
				fmt.Println(util.GetNatsImage())
				fmt.Println(util.GetDebuggerImage())
			}
		},
	}

	// Register flags
	cmd.Flags().Bool("ecn", false, "Get default package versions and images of all ECN components")

	return cmd
}
