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
	"github.com/eclipse-iofog/iofogctl/v2/internal/deploy"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "deploy",
		Example: `deploy -f platform.yaml`,
		Short:   "Deploy ioFog platform or components on existing infrastructure",
		Long: `Deploy ioFog platform or individual components on existing infrastructure.
The complete description of yaml file definition can be found at iofog.org`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			err = deploy.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed resources")
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
