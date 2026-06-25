package cmd

import (
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/all"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteAllCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Delete all resources within a namespace",
		Long: `Delete all resources within a namespace.

Tears down all components of an Edge Compute Network.

If you don't want to tear down the deployments but would like to free up the Namespace, use the disconnect command instead.`,
		Example: ex(`%[1]s delete all -n NAMESPACE`),
		Run: func(cmd *cobra.Command, args []string) {
			// Execute command
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)
			deleteNamespace, err := cmd.Flags().GetBool("delete-namespace")
			util.Check(err)
			err = delete.Execute(namespace, useDetached, force, deleteNamespace)
			util.Check(err)

			util.PrintSuccess("Successfully deleted all resources in namespace " + namespace)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion of Agents")
	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
