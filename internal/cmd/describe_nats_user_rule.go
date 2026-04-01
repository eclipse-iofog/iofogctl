package cmd

import (
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeNatsUserRuleCommand() *cobra.Command {
	output := natsOutputYAML
	cmd := &cobra.Command{
		Use:   "nats-user-rule NAME",
		Short: "Get detailed information about a NATS user rule",
		Long:  "Get detailed information about a NATS user rule.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			response, err := clt.ListNatsUserRules()
			util.Check(err)

			rule, err := findNatsRuleByName(response.Rules, args[0])
			util.Check(err)
			manifest := buildNatsRuleManifest("NatsUserRule", rule)
			table := [][]string{{"NAME"}, {rule.Name}}
			util.Check(printNatsOutput(manifest, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	return cmd
}
