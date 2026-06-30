package cmd

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/authgroup"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

const authGroupHelpLong = `Manage embedded auth groups (admin/sre only for create, rename, and delete).

System groups (admin, sre, developer, viewer) cannot be created or deleted; their names cannot be changed.
MfaRequired can be toggled on any group. Auth groups are unavailable when the Controller uses external OIDC.`

func newCreateAuthGroupCommand() *cobra.Command {
	mfaRequired := false
	cmd := &cobra.Command{
		Use:     "auth-group NAME",
		Short:   "Create a custom embedded auth group",
		Long:    authGroupHelpLong,
		Example: ex(`%[1]s create auth-group secops --mfa-required -n NAMESPACE`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)

			req := client.AuthGroupCreateRequest{Name: args[0]}
			if cmd.Flags().Changed("mfa-required") {
				value := mfaRequired
				req.MfaRequired = &value
			}

			group, err := clt.CreateAuthGroup(req)
			util.Check(authgroup.MapError(args[0], err))
			util.Check(printAuthGroupTable([]client.AuthGroupResponse{group}))
			util.PrintSuccess(fmt.Sprintf("Successfully created auth group %s", group.Name))
		},
	}
	cmd.Flags().BoolVar(&mfaRequired, "mfa-required", false, "Require TOTP for members of this group at login")
	return cmd
}

func newConfigureAuthGroupCommand() *cobra.Command {
	var newName string
	var mfaRequired bool
	cmd := &cobra.Command{
		Use:   "auth-group NAME",
		Short: "Update an embedded auth group",
		Long:  authGroupHelpLong,
		Example: ex(`%[1]s configure auth-group secops --name platform-ops -n NAMESPACE
%[1]s configure auth-group viewer --mfa-required=true -n NAMESPACE`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("mfa-required") {
				util.Check(util.NewInputError("at least one of --name or --mfa-required is required"))
			}

			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)

			req := client.AuthGroupUpdateRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = &newName
			}
			if cmd.Flags().Changed("mfa-required") {
				value := mfaRequired
				req.MfaRequired = &value
			}

			group, err := clt.UpdateAuthGroup(args[0], req)
			util.Check(authgroup.MapError(args[0], err))
			util.Check(printAuthGroupTable([]client.AuthGroupResponse{group}))
			util.PrintSuccess(fmt.Sprintf("Successfully configured auth group %s", group.Name))
		},
	}
	cmd.Flags().StringVar(&newName, "name", "", "New group name (custom groups only)")
	cmd.Flags().BoolVar(&mfaRequired, "mfa-required", false, "Require TOTP for members of this group at login")
	return cmd
}

func newDeleteAuthGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth-group NAME",
		Short:   "Delete a custom embedded auth group",
		Long:    authGroupHelpLong,
		Example: ex(`%[1]s delete auth-group secops -n NAMESPACE`),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			clt, err := clientutil.NewControllerClient(namespace)
			util.Check(err)

			err = clt.DeleteAuthGroup(args[0])
			util.Check(authgroup.MapError(args[0], err))
			util.PrintSuccess("Successfully deleted auth group " + args[0])
		},
	}
	return cmd
}

func printAuthGroupTable(groups []client.AuthGroupResponse) error {
	table := authgroup.Table(groups)
	return printWideTable(table)
}
