package cmd

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/configure"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newConfigureCommand() *cobra.Command {
	// Instantiate options
	var opt configure.Options

	cmd := &cobra.Command{
		Use:   "configure RESOURCE NAME",
		Short: ex("Configure %[1]s or ioFog resources"),
		Long: ex(`Configure %[1]s or ioFog resources

If you would like to replace the host value of Remote Controllers or Agents, you should delete and redeploy those resources.`),
		Example: ex(`%[1]s configure current-namespace NAME

%[1]s configure controller  NAME --user USER --key KEYFILE --port PORTNUM
                   controllers
                   agent
                   agents
				   controlplane

%[1]s configure controlplane --kube FILE`),
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				util.Check(util.NewInputError("Must specify a resource to configure"))
			}
			// Get resource type and name
			opt.ResourceType = args[0]
			if len(args) > 1 {
				opt.Name = args[1]
			} else if opt.ResourceType == "all" && opt.ResourceType != "agents" && opt.ResourceType != "controlplane" {
				util.Check(util.NewInputError("Must specify resource name if not configuring a group of resources"))
			}

			var err error

			// Get namespace option
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			opt.UseDetached, err = cmd.Flags().GetBool("detached")
			util.Check(err)

			// Get executor for configure command
			exe, err := configure.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess(fmt.Sprintf("Successfully configured %s %s", opt.ResourceType, opt.Name))
		},
	}
	cmd.Flags().StringVar(&opt.User, "user", "", "Username of remote host")
	cmd.Flags().StringVar(&opt.KeyFile, "key", "", "Path to private SSH key")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube", "", "Path to Kubernetes configuration file")
	cmd.Flags().StringVar(&opt.CAFile, "ca", "", "Path to PEM CA certificate for controller TLS (persisted to namespace config)")
	cmd.Flags().StringVar(&opt.CAB64, "ca-b64", "", "Base64-encoded PEM CA certificate for controller TLS (persisted to namespace config)")
	cmd.Flags().IntVar(&opt.Port, "port", 0, ex("Port number that %[1]s uses to SSH into remote hosts"))
	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
