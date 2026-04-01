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

package exec

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/internal/util/terminal"
	"github.com/eclipse-iofog/iofogctl/internal/util/websocket"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
	client    *client.Client
	msvc      *client.MicroserviceInfo
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	return a
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	util.SpinStart("Connecting Exec Session to Agent")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		msg := "%s\nFailed to get Agent by name: %s"
		return fmt.Errorf(msg, err.Error())
	}

	appName := fmt.Sprintf("system-%s", agent.Name)
	msvcName := fmt.Sprintf("debug-%s", agent.Name)

	exe.msvc, err = clt.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		// Check if error indicates application not found
		if strings.Contains(err.Error(), "Invalid application id") {
			// Try system application
			exe.msvc, err = clt.GetSystemMicroserviceByName(appName, msvcName)
			if err != nil {
				return err
			}
		} else {
			// Return other types of errors
			return err
		}
	}

	// Create WebSocket client
	wsClient := websocket.NewClient(exe.msvc.UUID)

	// Get controller endpoint
	controllerURL := clt.GetBaseURL()
	// Convert http(s):// to ws(s)://
	wsURL := strings.Replace(controllerURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = fmt.Sprintf("%s/microservices/system/exec/%s", wsURL, exe.msvc.UUID)

	// Set up headers
	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", clt.GetAccessToken()))
	util.SpinHandlePrompt()
	// Connect to WebSocket
	if err := wsClient.Connect(wsURL, headers); err != nil {
		util.SpinHandlePromptComplete()
		return util.NewError(fmt.Sprintf("failed to connect to WebSocket: %v", err))
	}

	// Create and start terminal
	term := terminal.NewTerminal(wsClient)

	// Check for initial connection error
	if err := wsClient.GetError(); err != nil {
		util.SpinHandlePromptComplete()
		formattedErr := formatWebSocketError(err)
		return util.NewError(formattedErr)
	}

	if err := term.Start(); err != nil {
		util.SpinHandlePromptComplete()
		formattedErr := formatWebSocketError(err)
		return util.NewError(formattedErr)
	}

	// Wait for terminal to finish
	<-wsClient.GetDone()
	msg := fmt.Sprintf("Successfully closed Agent %s Exec Session", exe.name)
	util.PrintSuccess(msg)

	// Check if there was an error
	if err := wsClient.GetError(); err != nil {
		formattedErr := formatWebSocketError(err)
		return util.NewError(formattedErr)
	}

	return nil
}
