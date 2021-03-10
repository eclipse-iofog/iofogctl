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
	attach "github.com/eclipse-iofog/iofogctl/v3/internal/attach/agent"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachAgentCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Attach an Agent to an existing Namespace",
		Long: `Attach a detached Agent to an existing Namespace.

The Agent will be provisioned with the Controller within the Namespace.`,
		Example: `iofogctl attach agent NAME`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(&opt)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully attached Agent " + opt.Name + " to namespace " + opt.Namespace)
		},
	}

	return cmd
}
