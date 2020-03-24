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
	"fmt"
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Open ECN Viewer",
		Run: func(cmd *cobra.Command, args []string) {
			// Get Control Plane
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			ns, err := config.GetNamespace(namespace)
			util.Check(err)
			controlPlane, err := ns.GetControlPlane()
			if err != nil || len(controlPlane.GetControllers()) == 0 {
				util.PrintError("You must deploy a Control Plane to a namespace to see an ECN Viewer")
				os.Exit(1)
			}
			ctrl := controlPlane.GetControllers()[0]
			url := util.Before(ctrl.GetEndpoint(), ":")
			if !strings.HasPrefix(url, "http") {
				url = "http://" + url
			}
			if util.IsLocalHost(util.After(url, "http://")) {
				url += ":" + iofog.ControllerHostECNViewerPortString
			}
			if err := browser.OpenURL(url); err != nil {
				util.PrintInfo("To see the ECN Viewer, open your browser and go to:\n")
				util.PrintInfo(fmt.Sprintf("%s\n", url))
			}
		},
	}
	return cmd
}
