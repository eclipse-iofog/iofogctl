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

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployApplicationCommand() *cobra.Command {
	// Instantiate options
	opt := &config.Application{}
	filename := ""

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "application",
		Example: `iofogctl deploy application -f application.yaml`,
		Short:   "Deploy ioFog application on existing infrastructure",
		Long: `Deploy ioFog application on existing infrastructure.
 
 A YAML resource definition file must be used to describe the application.
 
 The YAML application definition file should look like this :` + "\n```\n" +
			`name: "HealthcareWearableExample" # Application name
 microservices: # Array of mircroservices
		# First microservice
	 - name: "heart-rate-monitor" # Microservice name
		 agent:
			 name: "ioFog Agent" # Agent name to deploy the microservice on
			 config: # Optional - Required agent config for the microservice to run
				 bluetoothEnabled: true # this will install the iofog/restblue microservice
				 abstractedHardwareEnabled: false
		 images: # Microservice docker images
			 arm: "edgeworx/healthcare-heart-rate:arm-v1" # Image to deploy on agent of architecture ARM
			 x86: "edgeworx/healthcare-heart-rate:x86-v1" # Image to deploy on agent of architecture x86
		 roothostaccess: false # Does the docker container need host root access ?
		 ports: []
		 config: # Optional - Microservice configuration, free object
			 test_mode: true
			 data_label: "Anonymous Person"
		# Second microservice
	 - name: "heart-rate-viewer"
		 agent:
			 name: "ioFog Agent"
		 images:
			 arm: "edgeworx/healthcare-heart-rate-ui:arm"
			 # x86: "edgeworx/healthcare-heart-rate:x86-nano"
			 x86: "edgeworx/healthcare-heart-rate-ui:x86"
		 roothostaccess: false
		 ports: # Optional - Array of port mapping for the container
			 # The ui will be listening on port 80 (internal).
			 - external: 5000 # You will be able to access the ui on <AGENT_IP>:5000
				 internal: 80 # The ui is listening on port 80. Do not edit this.
		 volumes: # Optional - Array of volume mapping for the container
		   - hostdestination: /tmp/msvc # host volume
			 - containerdestination: /tmp # container volume
			 - accessmode: rw # access mode
		 env: # Optional - Array of environment variables for the container
			 - key: "BASE_URL"
				 value: "http://localhost:8080/data"
 routes:
	# Use this section to configure route between microservices
	# Use microservice name
	- from: "heart-rate-monitor"
		to: "heart-rate-viewer"
 ` + "\n```\n",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			// Unmarshall the input file
			err = util.UnmarshalYAML(filename, &opt)
			util.Check(err)

			// Get namespace
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Get executor for command
			executor, err := deployapplication.NewExecutor(namespace, opt)
			util.Check(err)

			// Execute the command
			err = executor.Execute()
			util.Check(err)

			util.PrintSuccess(fmt.Sprintf("Successfully deployed application %s to namespace %s", opt.Name, namespace))
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&filename, "file", "f", "", "YAML file containing application definition")

	return cmd
}
