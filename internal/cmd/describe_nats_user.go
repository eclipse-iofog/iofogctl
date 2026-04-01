package cmd

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeNatsUserCommand() *cobra.Command {
	output := natsOutputYAML
	jwtOnly := false
	cmd := &cobra.Command{
		Use:   "nats-user APP_NAME USER_NAME",
		Short: "Get detailed information about a NATS user",
		Long:  "Get detailed information about a NATS user.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			users, err := clt.ListNatsAccountUsers(args[0])
			util.Check(err)
			user, err := findNatsUserByName(users.Users, args[1])
			util.Check(err)

			if jwtOnly {
				printDecodedJWT(user.JWT, output)
				return
			}
			table := [][]string{
				{"NAME", "BEARER", "PUBLIC KEY", "JWT"},
				{user.Name, fmt.Sprintf("%t", user.IsBearer), user.PublicKey, user.JWT},
			}
			util.Check(printNatsOutput(user, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().BoolVar(&jwtOnly, "jwt", false, "Output decoded JWT payload only")
	return cmd
}
