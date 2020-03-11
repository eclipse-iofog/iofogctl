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
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get CLI application version",
		Run: func(cmd *cobra.Command, args []string) {
			util.PrintInfo("iofogctl Unified Cli\n")
			util.PrintInfo("Copyright (C) 2019, Edgeworx, Inc.\n")
			util.Print(util.GetVersion())
		},
	}
	return cmd
}
