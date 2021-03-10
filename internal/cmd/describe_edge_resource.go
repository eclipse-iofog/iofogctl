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
	"github.com/eclipse-iofog/iofogctl/v3/internal/describe"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
)

func newDescribeEdgeResourceCommand() *cobra.Command {
	opt := describe.Options{
		Resource: "edge-resource",
	}

	cmd := &cobra.Command{
		Use:     "edge-resource NAME VERSION",
		Short:   "Get detailed information about an Edge Resource",
		Long:    `Get detailed information about an Edge Resource.`,
		Example: `iofogctl describe edge-resource NAME VERSION`,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type and name
			var err error
			opt.Name = args[0]
			opt.Version = args[1]
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for describe command
			exe, err := describe.NewExecutor(&opt)
			util.Check(err)

			// Execute the command
			err = exe.Execute()
			util.Check(err)
		},
	}
	cmd.Flags().StringVarP(&opt.Filename, "output-file", "o", "", "YAML output file")

	return cmd
}
