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

type agentExecutor struct {
	namespace string
	name      string
	logConfig *LogTailConfig
}

func newAgentExecutor(namespace, name string, logConfig *LogTailConfig) *agentExecutor {
	exe := &agentExecutor{}
	exe.namespace = namespace
	exe.name = name
	exe.logConfig = logConfig
	return exe
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}

	// Get agent config
	baseAgent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		sdkCfg := localAgentSDKConfig(agent)
		lc, err := install.NewLocalContainerClient(install.LocalContainerEngineForHostOps(sdkCfg), sdkCfg)
		if err != nil {
			return err
		}
		stdout, stderr, err := lc.GetLogsByName(install.EdgeletContainerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		// Use WebSocket to stream logs from Controller
		util.SpinStart("Connecting to Agent Logs")

		// Init controller client
		clt, err := clientutil.NewControllerClient(exe.namespace)
		if err != nil {
			util.SpinHandlePromptComplete()
			return err
		}

		// Get agent UUID from controller
		agentInfo, err := clt.GetAgentByName(exe.name)
		if err != nil {
			util.SpinHandlePromptComplete()
			return fmt.Errorf("failed to get Agent by name: %s", err.Error())
		}

		// Create WebSocket client (using agent UUID as identifier)
		wsClient := ws.NewClient(agentInfo.UUID)

		// Get controller endpoint
		controllerURL := clt.GetBaseURL()
		// Convert http(s):// to ws(s)://
		wsURL := strings.Replace(controllerURL, "http://", "ws://", 1)
		wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
		wsURL = fmt.Sprintf("%s/iofog/%s/logs", wsURL, agentInfo.UUID)

		// Append query parameters from log config
		if exe.logConfig != nil {
			queryString := exe.logConfig.BuildQueryString()
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

func localAgentSDKConfig(agent *rsc.LocalAgent) *client.AgentConfiguration {
	if agent == nil || agent.Config == nil {
		return nil
	}
	return &agent.Config.AgentConfiguration
}
