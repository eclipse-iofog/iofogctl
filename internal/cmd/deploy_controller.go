/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
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
	"github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployControllerCommand() *cobra.Command {
	// Instantiate options
	var opt deploycontroller.Options

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "controller",
		Short:   "Deploy a Controller",
		Long:    `Deploy a Controller.`,
		Example: `iofogctl deploy controller -f controller.yaml`,
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute the command
			err = deploycontroller.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed Controllers to namespace " + opt.Namespace)
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controller")

	return cmd
}
