package cmd

import (
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	// Values accepted in resource type argument
	filename := ""
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Get detailed information of an existing resources",
		Long: `Get detailed information of an existing resources.
 
Most resources require a working Controller in the Namespace in order to be described.`,
	}

	// Add subcommands
	cmd.AddCommand(
		newDescribeControlPlaneCommand(),
		newDescribeControllerCommand(),
		newDescribeNamespaceCommand(),
		newDescribeAgentCommand(),
		newDescribeRegistryCommand(),
		newDescribeAgentConfigCommand(),
		newDescribeMicroserviceCommand(),
		newDescribeSystemMicroserviceCommand(),
		newDescribeApplicationCommand(),
		newDescribeApplicationTemplateCommand(),
		newDescribeVolumeCommand(),
		newDescribeEdgeResourceCommand(),
		newDescribeSecretCommand(),
		newDescribeConfigMapCommand(),
		newDescribeServiceCommand(),
		newDescribeVolumeMountCommand(),
		newDescribeCertificateCommand(),
		newDescribeRoleCommand(),
		newDescribeRoleBindingCommand(),
		newDescribeServiceAccountCommand(),
		newDescribeNatsAccountCommand(),
		newDescribeNatsUserCommand(),
		newDescribeNatsAccountRuleCommand(),
		newDescribeNatsUserRuleCommand(),
	)

	// Register Flags
	cmd.Flags().StringVarP(&filename, "output-file", "o", "", "YAML output file")

	return cmd
}
