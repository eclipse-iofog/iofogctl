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
	"github.com/eclipse-iofog/iofogctl/v3/internal/logs"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs RESOURCE NAME",
		Short: "Get log contents of deployed resource",
		Long:  `Get log contents of deployed resource`,
		Example: `iofogctl logs controller   NAME
              agent        NAME
              microservice NAME`,
		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get Resource type and name
			resource := args[0]
			name := args[1]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Instantiate logs executor
			exe, err := logs.NewExecutor(resource, namespace, name)
			util.Check(err)

			// Run the logs command
			err = exe.Execute()
			util.Check(err)
		},
	}

	return cmd
}
