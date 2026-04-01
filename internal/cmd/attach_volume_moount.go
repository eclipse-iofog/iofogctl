/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
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

	attach "github.com/eclipse-iofog/iofogctl/internal/attach/volumemount"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachVolumeMountCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:     "volume-mount NAME AGENT_NAME1 AGENT_NAME2",
		Short:   "Attach a Volume Mount to existing Agents",
		Long:    `Attach a Volume Mount to existing Agents.`,
		Example: `iofogctl attach volume-mount NAME AGENT_NAME1 AGENT_NAME2`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			opt.Agents = args[1:]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			msg := fmt.Sprintf("Successfully attached Volume Mount %s to Agents %s", opt.Name, strings.Join(opt.Agents, ", "))
			util.PrintSuccess(msg)
		},
	}

	return cmd
}
