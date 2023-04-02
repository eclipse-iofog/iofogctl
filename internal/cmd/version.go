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

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get CLI application version",
		Run: func(cmd *cobra.Command, args []string) {
			ecnFlag, err := cmd.Flags().GetBool("ecn")
			util.Check(err)
			util.PrintInfo("iofogctl - Copyright (C) 2019-2023, Edgeworx, Inc.\n")
			_ = util.Print(util.GetVersion())
			if ecnFlag {
				fmt.Println("")
				fmt.Println("controller@" + util.GetControllerVersion())
				fmt.Println("agent@" + util.GetAgentVersion())
				fmt.Println("")
				fmt.Println(util.GetControllerImage())
				fmt.Println(util.GetAgentImage())
				fmt.Println(util.GetOperatorImage())
				fmt.Println(util.GetKubeletImage())
				fmt.Println(util.GetPortManagerImage())
				fmt.Println(util.GetProxyImage())
				fmt.Println(util.GetRouterImage())
			}
		},
	}

	// Register flags
	cmd.Flags().Bool("ecn", false, "Get default package versions and images of all ECN components")

	return cmd
}
