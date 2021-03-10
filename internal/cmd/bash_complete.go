/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func newBashCompleteCommand(rootCmd *cobra.Command) *cobra.Command {
	// Find home directory.
	home, err := homedir.Dir()
	util.Check(err)
	configDir := home + "/.iofog/"
	err = os.MkdirAll(configDir, 0755)
	util.Check(err)
	cmd := &cobra.Command{
		Use:    "autocomplete SHELL",
		Hidden: true,
		Short:  "Generate bash autocomplete file",
		Long:   "Generate bash autocomplete file",
		Example: `iofogctl autocomplete bash
                      zsh`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch t := strings.ToLower(args[0]); t {
			case "bash":
				completionFilePath := configDir + "completion.bash.sh"
				err = rootCmd.GenBashCompletionFile(completionFilePath)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("%s generated", completionFilePath))
				util.PrintInfo(fmt.Sprintf("Run `source %s` to update your current session", completionFilePath))
				if runtime.GOOS == "darwin" {
					util.PrintInfo("If you have not done so yet, please install bash-completion using: `brew install bash-completion`")
				}
				util.PrintInfo(fmt.Sprintf("Add `source %s` to your bash profile to have it saved", completionFilePath))
			case "zsh":
				completionFilePath := configDir + "completion.bash.sh"
				err = rootCmd.GenZshCompletionFile(completionFilePath)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("%s generated", completionFilePath))
			default:
				util.Check(util.NewNotFoundError(fmt.Sprintf("%s shell not supported for autocompletion\n Supported shells are BASH and ZSH", t)))
			}
		},
	}
	return cmd
}
