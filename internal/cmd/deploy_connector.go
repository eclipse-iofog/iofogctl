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
	"github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployConnectorCommand() *cobra.Command {
	// Instantiate options
	var opt deployconnector.Options

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "connector",
		Short:   "Deploy a Connector",
		Long:    `Deploy a Connector.`,
		Example: `iofogctl deploy connector -f connector.yaml`,
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute the command
			err = deployconnector.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed Connectors to namespace " + opt.Namespace)
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Connector")

	return cmd
}
