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
	"github.com/eclipse-iofog/iofogctl/v3/internal/disconnect"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newDisconnectCommand() *cobra.Command {
	// Instantiate options
	opt := &disconnect.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from an ioFog cluster",
		Long: `Disconnect from an ioFog cluster.

This will remove all client-side information for this Namespace. The Namespace will itself be deleted.
Use the connect command to reconnect after a disconnect.
If you would like to uninstall the Control Plane and/or Agents, use the delete command instead.`,
		Example: `iofogctl disconnect -n NAMESPACE`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute the get command
			err = disconnect.Execute(opt)
			util.Check(err)
		},
	}

	return cmd
}
