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
	"github.com/eclipse-iofog/cli/pkg/root/deploy"
	"github.com/eclipse-iofog/cli/pkg/config"
)

//NewCommand export
func NewCommand() *cobra.Command {
	// Root command
	var cmd = &cobra.Command{
		Use:   "iofog",
		Short: "ioFog Unified Command Line Interface",
		Long: "ioFog Unified Command Line Interface",
	}

	// Initialize config filename
	cobra.OnInitialize(initConfig)

	// Global flags
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "CLI configuration file (default is $HOME/.cli.yaml)")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "Namespace to execute respective command within")

	// Register all commands
	cmd.AddCommand(get.NewCommand())
	cmd.AddCommand(deploy.NewCommand())

	return cmd
}

var configFilename string
func initConfig(){
	config.SetFile(configFilename)
}