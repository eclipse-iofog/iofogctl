package cmd

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeNatsAccountCommand() *cobra.Command {
	output := natsOutputYAML
	jwtOnly := false
	cmd := &cobra.Command{
		Use:   "nats-account APP_NAME",
		Short: "Get detailed information about a NATS account",
		Long:  "Get detailed information about a NATS account.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			account, err := clt.GetNatsAccount(args[0])
			util.Check(err)

			if jwtOnly {
				printDecodedJWT(account.JWT, output)
				return
			}
			table := [][]string{
				{"NAME", "APP ID", "SYSTEM", "PUBLIC KEY", "JWT"},
				{account.Name, fmt.Sprintf("%d", account.ApplicationID), fmt.Sprintf("%t", account.IsSystem), account.PublicKey, account.JWT},
			}
			util.Check(printNatsOutput(account, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().BoolVar(&jwtOnly, "jwt", false, "Output decoded JWT payload only")
	return cmd
}
