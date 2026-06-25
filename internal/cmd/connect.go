package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/connect"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newConnectCommand() *cobra.Command {
	// Instantiate options
	opt := connect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to an existing Control Plane",
		Long: ex(`Connect to an existing Control Plane.

This command must be executed within an empty or non-existent Namespace.
All resources provisioned with the corresponding Control Plane will become visible under the Namespace.
Visit %[2]s to view all YAML specifications usable with this command.`, util.GetCliDocsUrl()),
		Example: ex(`%[1]s connect -f controlplane.yaml

%[1]s connect --email EMAIL --pass PASSWORD --kube     FILE 
                 --email EMAIL --pass PASSWORD --ecn-addr ENDPOINT --name NAME

%[1]s connect --generate`),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			// Execute command
			err = connect.Execute(&opt)
			util.Check(err)

			if !opt.Generate {
				util.PrintSuccess("Successfully connected resources to namespace " + opt.Namespace)
			}
		},
	}
	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", pkg.flagDescYaml)
	cmd.Flags().StringVar(&opt.ControllerName, "name", "", "Name you would like to assign to Controller")
	cmd.Flags().StringVar(&opt.ControllerEndpoint, "ecn-addr", "", "URL of Edge Compute Network to connect to")
	cmd.Flags().StringVar(&opt.KubeConfig, "kube", "", "Kubernetes config file. Typically ~/.kube/config")
	cmd.Flags().StringVar(&opt.IofogUserEmail, "email", "", "ioFog user email address")
	cmd.Flags().StringVar(&opt.IofogUserPass, "pass", "", "ioFog user password")
	cmd.Flags().BoolVar(&opt.OverwriteNamespace, "force", false, "Overwrite existing Namespace")
	cmd.Flags().BoolVar(&opt.Generate, "generate", false, "Generate a connection string that can be used to connect to this ECN")
	cmd.Flags().BoolVar(&opt.Base64Encoded, "b64", false, "Indicate whether input password (--pass) is base64 encoded or not")
	cmd.Flags().StringVar(&opt.CAFile, "ca", "", "Path to PEM CA certificate for controller TLS (persisted to namespace config)")
	cmd.Flags().StringVar(&opt.CAB64, "ca-b64", "", "Base64-encoded PEM CA certificate for controller TLS (persisted to namespace config)")

	return cmd
}
