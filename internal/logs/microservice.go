/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package logs

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	ws "github.com/eclipse-iofog/iofogctl/internal/util/websocket"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteMicroserviceExecutor struct {
	namespace string
	name      string
	logConfig *LogTailConfig
}

func newRemoteMicroserviceExecutor(namespace, name string, logConfig *LogTailConfig) *remoteMicroserviceExecutor {
	m := &remoteMicroserviceExecutor{}
	m.namespace = namespace
	m.name = name
	m.logConfig = logConfig
	return m
}

func (ms *remoteMicroserviceExecutor) GetName() string {
	return ms.name
}

func (ms *remoteMicroserviceExecutor) Execute() error {
	// Get image name of the microservice and details of the Agent its deployed on
	baseAgent, msvc, isSystem, err := getAgentAndMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}

	if msvc.Status.Status != "RUNNING" {
		return util.NewError("The microservice is not currently running")
	}

	switch baseAgent.(type) {
	case *rsc.LocalAgent:
		lc, err := install.NewLocalContainerClient()
		if err != nil {
			return err
		}
		containerName := "iofog_" + msvc.UUID
		stdout, stderr, err := lc.GetLogsByName(containerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		// Use WebSocket to stream logs from Controller
		util.SpinStart("Connecting to Microservice Logs")

		// Init controller client
		clt, err := clientutil.NewControllerClient(ms.namespace)
		if err != nil {
			util.SpinHandlePromptComplete()
			return err
		}

		// Create WebSocket client (using microservice UUID)
		wsClient := ws.NewClient(msvc.UUID)

		// Get controller endpoint
		controllerURL := clt.GetBaseURL()
		// Convert http(s):// to ws(s)://
		wsURL := strings.Replace(controllerURL, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		if isSystem {
			wsURL = fmt.Sprintf("%s/microservices/system/%s/logs", wsURL, msvc.UUID)
		} else {
			wsURL = fmt.Sprintf("%s/microservices/%s/logs", wsURL, msvc.UUID)
		}

		// Append query parameters from log config
		if ms.logConfig != nil {
			queryString := ms.logConfig.BuildQueryString()
			if queryString != "" {
				wsURL = fmt.Sprintf("%s?%s", wsURL, queryString)
			}
		}

		// Set up headers
		headers := http.Header{}
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", clt.GetAccessToken()))
		util.SpinHandlePrompt()
		// Connect to WebSocket
		if err := wsClient.Connect(wsURL, headers); err != nil {
			util.SpinHandlePromptComplete()
			return util.NewError(fmt.Sprintf("failed to connect to WebSocket: %v", err))
		}

		// Create and start log stream
		logStream := NewLogStream(wsClient)

		// Check for initial connection error
		if err := wsClient.GetError(); err != nil {
			util.SpinHandlePromptComplete()
			formattedErr := formatWebSocketError(err)
			return util.NewError(formattedErr)
		}

		if err := logStream.Start(); err != nil {
			util.SpinHandlePromptComplete()
			formattedErr := formatWebSocketError(err)
			return util.NewError(formattedErr)
		}

		// Wait for stream to finish
		<-wsClient.GetDone()
		util.SpinHandlePromptComplete()
	}

	return nil
}

// func (ms *remoteMicroserviceExecutor) runDockerCommand(cmd string, ssh *util.SecureShellClient) (stdout bytes.Buffer, err error) {
// 	stdout, err = ssh.Run(cmd)
// 	if err != nil {
// 		if !strings.Contains(strings.ToLower(err.Error()), "permission denied") {
// 			return
// 		}
// 		// Retry with sudo
// 		cmd = strings.Replace(cmd, "docker", "sudo docker", -1)

// 		stdout, err = ssh.Run(cmd)
// 		if err != nil {
// 			return
// 		}
// 	}
// 	return
// }

func getAgentAndMicroservice(namespace, msvcFQName string) (agent rsc.Agent, msvc client.MicroserviceInfo, isSystem bool, err error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}

	ctrlClient, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return
	}

	appName, msvcName, err := clientutil.ParseFQName(msvcFQName, "Microservice")
	if err != nil {
		return agent, msvc, false, err
	}

	// Get microservice details from Controller
	msvcPtr, err := ctrlClient.GetMicroserviceByName(appName, msvcName)
	isSystem = false
	if err != nil {
		// Check if error indicates application not found
		if strings.Contains(err.Error(), "Invalid application id") {
			// Try system application
			msvcPtr, err = ctrlClient.GetSystemMicroserviceByName(appName, msvcName)
			if err != nil {
				return
			}
			isSystem = true
		} else {
			// Return other types of errors
			return
		}
	}

	msvc = *msvcPtr

	// Get Agent running the microservice
	agentResponse, err := ctrlClient.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	agent, err = ns.GetAgent(agentResponse.Name)
	if err != nil {
		return
	}
	return agent, msvc, isSystem, nil
}
