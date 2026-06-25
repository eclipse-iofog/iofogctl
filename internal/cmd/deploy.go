package cmd

import (
	"errors"

	"github.com/eclipse-iofog/iofogctl/internal/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use: "deploy",
		Example: `deploy -f ecn.yaml
          application-template.yaml
          application.yaml
          microservice.yaml
          catalog.yaml
          volume.yaml
          route.yaml
          secret.yaml
          configmap.yaml
          service.yaml
          volume-mount.yaml`,

		Args:  cobra.ExactArgs(0),
		Short: "Deploy Edge Compute Network components on existing infrastructure",
		Long: `Deploy Edge Compute Network components on existing infrastructure.
Visit iofog.org to view all YAML specifications usable with this command.`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Check file
			if opt.InputFile == "" {
				util.Check(errors.New("provided empty value for input file via the -f flag"))
			}

			// Execute command
			err = deploy.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed resources")
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", pkg.flagDescYaml)
	cmd.Flags().BoolVar(&opt.NoCache, "no-cache", false, "Disable caching for OfflineImage images after download")
	cmd.Flags().IntVar(&opt.TransferPool, "transfer-pool", 2, "Maximum number of concurrent OfflineImage transfers")

	return cmd
}
