package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

const (
	natsOutputYAML = "yaml"
	natsOutputJSON = "json"
	natsOutputWide = "wide"
)

func newNatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nats",
		Short: "Manage NATS resources",
		Long:  "Manage NATS-specific operations exposed by Controller APIs. Use get/describe/deploy/delete for CRUD-style NATS resources.",
		Example: ex(`%[1]s nats operator describe
%[1]s nats accounts ensure my-app --nats-rule default-account-rule
%[1]s nats users create my-app service-user
%[1]s nats users creds my-app service-user`),
	}

	cmd.AddCommand(
		newNatsOperatorCommand(),
		newNatsAccountsCommand(),
		newNatsUsersCommand(),
	)

	return cmd
}

func newNatsOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "NATS operator operations",
		Long:  "Inspect NATS operator metadata and JWT information.",
	}
	cmd.AddCommand(newNatsOperatorDescribeCommand())
	return cmd
}

func newNatsOperatorDescribeCommand() *cobra.Command {
	output := natsOutputYAML
	jwtOnly := false
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe NATS operator",
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			operator, err := clt.GetNatsOperator()
			util.Check(err)

			if jwtOnly {
				printDecodedJWT(operator.JWT, output)
				return
			}

			wide := [][]string{
				{"NAME", "PUBLIC KEY", "JWT"},
				{operator.Name, operator.PublicKey, operator.JWT},
			}
			util.Check(printNatsOutput(operator, output, wide))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().BoolVar(&jwtOnly, "jwt", false, "Output decoded JWT payload only")
	return cmd
}

func newNatsAccountsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "accounts",
		Aliases: []string{"account"},
		Short:   "NATS account operations",
		Long:    "NATS-specific account actions for applications.",
		Example: ex(`%[1]s nats accounts ensure my-application --nats-rule default-account-rule`),
	}
	cmd.AddCommand(
		newNatsAccountsEnsureCommand(),
	)
	return cmd
}

func newNatsAccountsEnsureCommand() *cobra.Command {
	output := natsOutputYAML
	natsRule := ""
	cmd := &cobra.Command{
		Use:   "ensure APP_NAME",
		Short: "Ensure NATS account for application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			req := &client.NatsEnsureAccountRequest{}
			if natsRule != "" {
				req.NatsRule = natsRule
			}

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			account, err := clt.EnsureNatsAccount(args[0], req)
			util.Check(err)

			table := [][]string{
				{"NAME", "APP ID", "SYSTEM", "PUBLIC KEY"},
				{account.Name, fmt.Sprintf("%d", account.ApplicationID), fmt.Sprintf("%t", account.IsSystem), account.PublicKey},
			}
			util.Check(printNatsOutput(account, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().StringVar(&natsRule, "nats-rule", "", "NATS account rule name")
	return cmd
}

func newNatsUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "NATS user operations",
		Long:    "NATS-specific user actions such as create/delete and creds retrieval.",
		Example: ex(`%[1]s nats users create my-application service-user
%[1]s nats users creds my-application service-user
%[1]s nats users creds my-application service-user -o ./service-user.creds`),
	}
	cmd.AddCommand(
		newNatsUsersCreateCommand(),
		newNatsUsersDeleteCommand(),
		newNatsUsersCreateMqttBearerCommand(),
		newNatsUsersDeleteMqttBearerCommand(),
		newNatsUsersCredsCommand(),
	)
	return cmd
}

func newNatsUsersCreateCommand() *cobra.Command {
	output := natsOutputYAML
	var expiresIn int64
	natsRule := ""
	cmd := &cobra.Command{
		Use:   "create APP_NAME USER_NAME",
		Short: "Create NATS user under an application account",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			req := &client.NatsCreateUserRequest{
				Name: args[1],
			}
			if expiresIn > 0 {
				req.ExpiresIn = &expiresIn
			}
			if natsRule != "" {
				req.NatsRule = natsRule
			}

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			user, err := clt.CreateNatsUser(args[0], req)
			util.Check(err)

			table := [][]string{
				{"NAME", "BEARER", "PUBLIC KEY"},
				{user.Name, fmt.Sprintf("%t", user.IsBearer), user.PublicKey},
			}
			util.Check(printNatsOutput(user, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().Int64Var(&expiresIn, "expires-in", 0, "Expiry in seconds")
	cmd.Flags().StringVar(&natsRule, "nats-rule", "", "NATS user rule name")
	return cmd
}

func newNatsUsersDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete APP_NAME USER_NAME",
		Short: "Delete NATS user from an application account",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			err = clt.DeleteNatsUser(args[0], args[1])
			util.Check(err)
			util.PrintSuccess(fmt.Sprintf("Deleted NATS user %s/%s", args[0], args[1]))
		},
	}
	return cmd
}

func newNatsUsersCreateMqttBearerCommand() *cobra.Command {
	output := natsOutputYAML
	var expiresIn int64
	natsRule := ""
	cmd := &cobra.Command{
		Use:   "create-mqtt-bearer APP_NAME USER_NAME",
		Short: "Create MQTT bearer NATS user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			req := &client.NatsCreateMqttBearerRequest{
				Name: args[1],
			}
			if expiresIn > 0 {
				req.ExpiresIn = &expiresIn
			}
			if natsRule != "" {
				req.NatsRule = natsRule
			}

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			user, err := clt.CreateNatsMqttBearer(args[0], req)
			util.Check(err)

			table := [][]string{
				{"NAME", "PUBLIC KEY", "JWT"},
				{user.Name, user.PublicKey, user.JWT},
			}
			util.Check(printNatsOutput(user, output, table))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "", natsOutputYAML, "Output format: yaml|json|wide")
	cmd.Flags().Int64Var(&expiresIn, "expires-in", 0, "Expiry in seconds")
	cmd.Flags().StringVar(&natsRule, "nats-rule", "", "NATS user rule name")
	return cmd
}

func newNatsUsersDeleteMqttBearerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-mqtt-bearer APP_NAME USER_NAME",
		Short: "Delete MQTT bearer NATS user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			err = clt.DeleteNatsMqttBearer(args[0], args[1])
			util.Check(err)
			util.PrintSuccess(fmt.Sprintf("Deleted MQTT bearer user %s/%s", args[0], args[1]))
		},
	}
	return cmd
}

func newNatsUsersCredsCommand() *cobra.Command {
	outputFile := ""
	cmd := &cobra.Command{
		Use:     "creds APP_NAME USER_NAME",
		Aliases: []string{"get-creds"},
		Short:   "Fetch NATS creds",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)
			credsResponse, err := clt.GetNatsUserCreds(args[0], args[1])
			util.Check(err)

			decoded, err := base64.StdEncoding.DecodeString(credsResponse.CredsBase64)
			util.Check(err)
			if strings.TrimSpace(outputFile) == "" {
				_, err = os.Stdout.Write(decoded)
				util.Check(err)
				return
			}
			err = os.WriteFile(outputFile, decoded, 0600)
			util.Check(err)
			util.PrintSuccess(fmt.Sprintf("Saved NATS creds to %s", outputFile))
		},
	}
	cmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Destination creds file path (always overwritten)")
	return cmd
}

func printNatsOutput(obj interface{}, output string, wide [][]string) error {
	switch strings.ToLower(output) {
	case natsOutputYAML:
		return util.Print(obj)
	case natsOutputJSON:
		j, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, string(j))
		return err
	case natsOutputWide:
		if len(wide) == 0 {
			return nil
		}
		return printWideTable(wide)
	default:
		return util.NewInputError("invalid --output, use yaml|json|wide")
	}
}

func printWideTable(table [][]string) error {
	writer := tabwriter.NewWriter(os.Stdout, 16, 8, 1, '\t', 0)
	defer writer.Flush()
	for _, row := range table {
		for _, col := range row {
			if _, err := fmt.Fprintf(writer, "%s\t", col); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}

func findNatsUserByName(users []client.NatsUserInfo, name string) (client.NatsUserInfo, error) {
	for _, user := range users {
		if user.Name == name {
			return user, nil
		}
	}
	return client.NatsUserInfo{}, util.NewNotFoundError(fmt.Sprintf("NATS user not found: %s", name))
}

func printDecodedJWT(token, output string) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		util.Check(util.NewInputError("invalid JWT token"))
		return
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	util.Check(err)

	var payload interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	util.Check(err)

	util.Check(printNatsOutput(payload, output, nil))
}
