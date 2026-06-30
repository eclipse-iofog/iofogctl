package cmd

import (
	"errors"

	"github.com/eclipse-iofog/iofogctl/internal/delete"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	// Instantiate options
	opt := &delete.Options{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing ioFog resource",
		Long:  `Delete an existing ioFog resource.`,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			opt.DeleteNamespace, err = cmd.Flags().GetBool("delete-namespace")
			util.Check(err)

			// Check file
			if opt.InputFile == "" {
				util.Check(errors.New("provided empty value for input file via the -f flag"))
			}

			// Execute command
			err = delete.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deleted resources from namespace " + opt.Namespace)
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newDeleteNamespaceCommand(),
		newDeleteControllerCommand(),
		newDeleteAgentCommand(),
		newDeleteAllCommand(),
		newDeleteApplicationCommand(),
		newDeleteApplicationTemplateCommand(),
		newDeleteCatalogItemCommand(),
		newDeleteRegistryCommand(),
		newDeleteMicroserviceCommand(),
		newDeleteVolumeCommand(),
		newDeleteSecretCommand(),
		newDeleteConfigMapCommand(),
		newDeleteRoleCommand(),
		newDeleteRoleBindingCommand(),
		newDeleteServiceAccountCommand(),
		newDeleteNatsAccountRuleCommand(),
		newDeleteNatsUserRuleCommand(),
		newDeleteServiceCommand(),
		newDeleteVolumeMountCommand(),
		newDeleteCertificateCommand(),
		newDeleteAuthGroupCommand(),
	)

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", pkg.flagDescYaml)
	cmd.PersistentFlags().BoolVar(&opt.DeleteNamespace, "delete-namespace", false, `Also delete the Kubernetes namespace (never deletes "default")`)

	return cmd
}
