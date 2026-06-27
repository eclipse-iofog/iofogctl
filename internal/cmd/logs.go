package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/logs"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs RESOURCE NAME",
		Short: "Get log contents of deployed resource",
		Long:  `Get log contents of deployed resource`,
		Example: ex(`%[1]s logs controller   NAME
              agent        NAME
              microservice AppName/MsvcName`),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get Resource type and name
			resource := args[0]
			name := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Parse log tail configuration flags
			tail, err := cmd.Flags().GetInt("tail")
			util.Check(err)
			follow, err := cmd.Flags().GetBool("follow")
			util.Check(err)
			since, err := cmd.Flags().GetString("since")
			util.Check(err)
			until, err := cmd.Flags().GetString("until")
			util.Check(err)

			// Create log tail config
			logConfig := &logs.LogTailConfig{
				Tail:   tail,
				Follow: follow,
				Since:  since,
				Until:  until,
			}

			// Validate config
			err = logConfig.Validate()
			util.Check(err)

			// Instantiate logs executor
			exe, err := logs.NewExecutor(resource, namespace, name, logConfig)
			util.Check(err)

			// Run the logs command
			err = exe.Execute()
			util.Check(err)
		},
	}

	// Add flags for log tail configuration
	cmd.Flags().Int("tail", 100, "Number of lines to tail (range: 1-5000)")
	cmd.Flags().Bool("follow", true, "Follow log output")
	cmd.Flags().String("since", "", "Start time in ISO 8601 format (e.g., 2024-01-01T00:00:00Z)")
	cmd.Flags().String("until", "", "End time in ISO 8601 format (e.g., 2024-01-02T00:00:00Z)")

	return cmd
}
