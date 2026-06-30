package cmd

import (
	"errors"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/get"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newGetCommand() *cobra.Command {
	validResources := []string{
		"all",
		"namespaces",
		"controllers",
		"agents",
		"application-templates",
		"applications",
		"system-applications",
		"microservices",
		"system-microservices",
		"catalog",
		"registries",
		"volumes",
		"secrets",
		"configmaps",
		"services",
		"volume-mounts",
		"certificates",
		"roles",
		"rolebindings",
		"serviceaccounts",
		"nats-accounts",
		"nats-users",
		"nats-account-rules",
		"nats-user-rules",
		"auth-groups",
		"auth-group",
	}
	cmd := &cobra.Command{
		Use:   "get RESOURCE [NAME]",
		Short: "Get information of existing resources",
		Long: `Get information of existing resources.

Resources like Agents will require a working Controller in the namespace to display all information.`,
		Example: ex(`%[1]s get all
             namespaces
             controllers
             agents
             application-templates
             applications
             system-applications
             microservices
             system-microservices
             catalog
             registries
             volumes
             secrets
             configmaps
             services
             volume-mounts
             certificates
             roles
             rolebindings
             serviceaccounts
             nats-accounts
             nats-users
             nats-account-rules
             nats-user-rules
             auth-groups
             auth-group NAME`),
		ValidArgs: validResources,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return util.NewInputError("expected 'get RESOURCE' or 'get auth-group NAME'")
			}
			if args[0] == "auth-group" {
				if len(args) != 2 {
					return util.NewInputError("get auth-group requires a name")
				}
				return nil
			}
			if len(args) != 1 {
				return util.NewInputError("get " + args[0] + " does not accept a name argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			resourceName := ""
			if len(args) == 2 {
				resourceName = args[1]
			}
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			showDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			// TODO: Break out resources as subcommands to avoid this kind of logic and improve --help accuracy
			if showDetached && resource != "agents" {
				err = errors.New("can only use --detached flag with Agents")
				util.Check(err)
			}

			if showDetached && namespace != config.GetDefaultNamespaceName() {
				util.PrintNotify("You are requesting detached resources, Namespace will be ignored.")
			}

			// Get executor for get command
			exe, err := get.NewExecutor(resource, namespace, showDetached, resourceName)
			util.Check(err)

			// Execute the get command
			err = exe.Execute()
			util.Check(err)
		},
	}

	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
