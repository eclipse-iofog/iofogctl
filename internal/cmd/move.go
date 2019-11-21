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
	"github.com/spf13/cobra"
)

func newMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get RESOURCE",
		Short:   "Get information of existing resources",
		Long:    `Get information of existing resources.`,
		Example: `iofogctl move microservice NAME AGENT_NAME`,
	}

	cmd.AddCommand(
		newMoveMicroserviceCommand(),
	)

	return cmd
}
