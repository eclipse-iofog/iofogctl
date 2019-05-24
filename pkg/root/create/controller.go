package deploy

import (
	"fmt"
	"github.com/spf13/cobra"
)

var controllerCommand = &cobra.Command{
	Use:   "controller",
	Short: "Brief",
	Long:  "Long",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy controller")
	},
}