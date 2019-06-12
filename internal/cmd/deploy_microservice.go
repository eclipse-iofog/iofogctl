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
	deploy "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployMicroserviceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microservice NAME [AGENTNAME]",
		Short: "Deploy a Microservice",
		Long:  `Deploy a Microservice`,
		Example: `iofogctl deploy microservice NAME
iofogctl deploy microservice NAME AGENTNAME`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			microservice := deploy.New()
			err = microservice.Execute(namespace, name)
			util.Check(err)
		},
	}

	return cmd
}
