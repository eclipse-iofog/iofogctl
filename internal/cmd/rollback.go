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
	"strings"

	"github.com/eclipse-iofog/iofogctl/v3/internal/rollback"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newRollbackCommand() *cobra.Command {
	// Instantiate options
	var opt rollback.Options

	cmd := &cobra.Command{
		Use:     "rollback RESOURCE NAME",
		Short:   "Rollback ioFog resources",
		Long:    `Rollback ioFog resources to latest versions available.`,
		Example: `iofogctl rollback agent NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			opt.ResourceType = args[0]
			opt.Name = args[1]

			var err error
			// Get namespace option
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for rollback command
			exe, err := rollback.NewExecutor(opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess(fmt.Sprintf("Succesfully scheduled rollback for %s %s", strings.Title(opt.ResourceType), opt.Name))
		},
	}

	return cmd
}
