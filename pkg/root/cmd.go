// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package root

import (
	"github.com/spf13/cobra"
	"github.com/eclipse-iofog/cli/pkg/root/get"
	"github.com/eclipse-iofog/cli/pkg/config"
)

//NewCommand export
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cli",
		Short: "A brief description of your application",
		Long: "A brief description of your application",
		//	Run: func(cmd *cobra.Command, args []string) { },
	}
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "config file (default is $HOME/.cli.yaml)")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "--namespace default")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cmd.AddCommand(get.NewCommand())
	return cmd
}

var configFilename string
func initConfig(){
	config.SetFile(configFilename)
}