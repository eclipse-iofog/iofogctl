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
	"io/ioutil"
	"os"
	"path"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update iofogctl config folder",
		Long:    "Update iofogctl config folder",
		Example: `iofogctl update`,
		Run: func(cmd *cobra.Command, args []string) {
			// Previous config structure
			type OldConfig struct {
				Namespaces []config.Namespace `yaml:"namespaces"`
			}

			// Get config folder
			configFolder, err := cmd.Flags().GetString("config")
			util.Check(err)

			if configFolder == "" {
				configFolder = "~/.iofog/"
			}

			configFolder, err = util.FormatPath(configFolder)
			util.Check(err)

			// Get config files
			configFileName := path.Join(configFolder, "config.yaml")
			configSaveFileName := path.Join(configFolder, "config.yaml.save")

			// Create namespaces folder
			namespaceDirectory := path.Join(configFolder, "namespaces")
			err = os.MkdirAll(namespaceDirectory, 0755)
			util.Check(err)

			// Read previous config
			r, err := ioutil.ReadFile(configFileName)
			util.Check(err)

			oldConfig := OldConfig{}
			newConfig := config.Configuration{DefaultNamespace: "default"}
			err = yaml.UnmarshalStrict(r, &oldConfig)
			if err != nil {
				if err2 := yaml.UnmarshalStrict(r, &newConfig); err2 != nil {
					util.Check(err)
				}
				util.PrintNotify(fmt.Sprintf("Your config file is up-to-date.\nThe previous config file has been saved under %s\nIf you encounter any issue, please contact us on slack, github or forum\n", configSaveFileName))
				return
			}

			// Map old config to new confi file system
			for _, ns := range oldConfig.Namespaces {
				// Add namespace to list
				newConfig.Namespaces = append(newConfig.Namespaces, ns.Name)

				// Write namespace config file
				bytes, err := yaml.Marshal(ns)
				util.Check(err)
				configFile := path.Join(namespaceDirectory, ns.Name+".yaml")
				err = ioutil.WriteFile(configFile, bytes, 0644)
				util.Check(err)
			}

			// Write old config save file
			err = ioutil.WriteFile(configSaveFileName, r, 0644)
			util.Check(err)

			// Write new config file
			bytes, err := yaml.Marshal(newConfig)
			util.Check(err)
			err = ioutil.WriteFile(configFileName, bytes, 0644)
			util.Check(err)

			util.PrintSuccess("Your config file has successfully been updated")
		},
	}

	return cmd
}
