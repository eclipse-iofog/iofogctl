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
	delete "github.com/eclipse-iofog/iofogctl/v2/internal/delete/namespace"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newDeleteNamespaceCommand() *cobra.Command {
	force := false
	soft := false
	cmd := &cobra.Command{
		Use:   "namespace NAME",
		Short: "Delete a Namespace",
		Long: `Delete a Namespace.

The namespace must not have any resources within it.`,
		Example: `iofogctl delete namespace NAME`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get microservice name
			name := args[0]

			// Execute command
			err := delete.Execute(name, force, soft)
			util.Check(err)

			util.PrintSuccess("Successfully deleted namespace " + name)
		},
	}

	cmd.Flags().BoolVar(&soft, "soft", false, "Don't delete ioFog stack from remote hosts")
	cmd.Flags().BoolVar(&force, "force", false, "Force deletion of all resources within the namespace")

	return cmd
}
