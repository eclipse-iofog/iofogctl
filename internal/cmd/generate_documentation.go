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
	"path"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenerateDocumentationCommand(rootCmd *cobra.Command) *cobra.Command {
	// Find home directory.
	home, err := homedir.Dir()
	var docDir string
	util.Check(err)
	cmd := &cobra.Command{
		Use:    "documentation TYPE",
		Hidden: true,
		Short:  "Generate iofogctl documentation",
		Long:   "Generate iofogctl documentation as markdown or man page",
		Example: `iofogctl documentation md
		 iofogctl documentation man`,
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if docDir == "" {
				docDir := home + "/.iofog/docs/"
				os.MkdirAll(docDir, 0755)
			}
			switch t := strings.ToLower(args[0]); t {
			case "md":
				mdDir := path.Join(docDir, "md/")
				os.MkdirAll(mdDir, 0755)
				err := doc.GenMarkdownTree(rootCmd, mdDir)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("markdown documentation generated at %s", mdDir))
			case "man":
				manDir := path.Join(docDir, "man/")
				os.MkdirAll(manDir, 0755)
				header := &doc.GenManHeader{
					Title:   "iofogctl",
					Section: "1",
				}
				err := doc.GenManTree(rootCmd, header, manDir)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("man documentation generated at %s", manDir))
			default:
				util.Check(util.NewNotFoundError(fmt.Sprintf("%s documentation format not supported for documentation generation\n Supported types are MAN and MD", t)))
			}
		},
	}

	cmd.Flags().StringVarP(&docDir, "output-dir", "o", "", "Output dir path")
	return cmd
}
