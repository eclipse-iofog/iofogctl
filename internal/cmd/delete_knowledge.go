package cmd

import (
	deleteknowledge "github.com/eclipse-iofog/iofogctl/internal/delete/knowledge"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteKnowledgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "knowledge NAME",
		Short:   "Delete Knowledge",
		Long:    `Delete Knowledge from the Controller.`,
		Example: ex(`%[1]s delete knowledge NAME`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := deleteknowledge.NewExecutor(namespace, name)
			util.Check(err)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully deleted knowledge " + name)
		},
	}

	return cmd
}
