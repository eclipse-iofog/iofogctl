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
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployMicroserviceCommand() *cobra.Command {
	// Instantiate options
	var opt deploymicroservice.Options

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "microservice",
		Example: `iofogctl deploy microservice -f microservice.yaml`,
		Short:   "Deploy ioFog microservice on existing infrastructure",
		Long: `Deploy ioFog microservice on existing infrastructure.
	
	A YAML resource definition file must be used to describe the application.
	
	The YAML microservice definition file should look like this :` + "\n```\n" +
			`name: "heart-rate-monitor" # Microservice name
	 agent:
		 name: "ioFog Agent" # Agent name to deploy the microservice on
		 config: # Optional - Required agent config for the microservice to run
			 bluetoothEnabled: true # this will install the iofog/restblue microservice
			 abstractedHardwareEnabled: false
	 images: # Microservice docker images
		 catalogid: 106 # Optional existing catalog id
		 arm: "edgeworx/healthcare-heart-rate:arm-v1" # Image to deploy on agent of architecture ARM
		 x86: "edgeworx/healthcare-heart-rate:x86-v1" # Image to deploy on agent of architecture x86
		 registry: remote # Optional. Either remote or local. Defines if the images are pulled from a remote image repository or if they are locally on the agent
	 roothostaccess: false # Does the docker container need host root access ?
	 application: HealthcareWearableExample # Application name
	 config: # Optional - Microservice configuration, free object
		 test_mode: true
		 data_label: "Anonymous Person"
	 ports: # Optional - Array of port mapping for the container
		 - external: 5000 # You will be able to access the ui on <AGENT_IP>:5000
			 internal: 80 # The ui is listening on port 80. Do not edit this.
	 volumes: # Optional - Array of volume mapping for the container
		 - hostdestination: /tmp/msvc # host volume
			 containerdestination: /tmp # container volume
			 accessmode: rw # access mode
	 env: # Optional - Array of environment variables for the container
		 - key: "BASE_URL"
			 value: "http://localhost:8080/data"
	 routes: # Optional - Array of destination microservice name
	   - heart-rate-viewer
	` + "\n```\n",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute the command
			err = deploymicroservice.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed Microservice to namespace " + opt.Namespace)
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing microservice definition")

	return cmd
}
