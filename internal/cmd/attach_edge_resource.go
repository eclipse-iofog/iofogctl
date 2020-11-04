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
	attach "github.com/eclipse-iofog/iofogctl/v2/internal/attach/edge-resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachEdgeResourceCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "edge-resource NAME/VERSION AGENT_NAME",
		Short:   "Attach an Edge Resource to an existing Agent",
		Long:    `Attach an Edge Resource to an existing Agent.`,
		Example: `iofogctl attach edge-resource NAME/VERSION AGENT_NAME`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.NameVersion = args[0]
			opt.Agent = args[1]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully attached EdgeResource " + opt.NameVersion + " to namespace " + opt.Namespace)
		},
	}

	return cmd
}
