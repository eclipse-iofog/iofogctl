package deletecontroller

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	"github.com/spf13/cobra"
)

// NewCommand export
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller name",
		Short: "Delete a Controller",
		Long: `Delete a Controller`,
		Example: `iofog delete controller my_controller_name`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe, err := getExecutor(namespace, name)
			util.Check(err)	

			err = exe.execute()
			util.Check(err)
		},
	}

	return cmd
}