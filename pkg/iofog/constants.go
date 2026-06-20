package iofog

import "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"

// String and numeric values of TCP ports used across ioFog
const (
	ControllerPort       = client.ControllerPort
	ControllerPortString = client.ControllerPortString

	ControllerHostECNViewerPort       = 8008
	ControllerHostECNViewerPortString = "8008"

	DefaultHTTPPort       = 8008
	DefaultHTTPPortString = "8008"

	// VanillaRemoteAgentName string = "0-controlplane"
	VanillaRouterAgentName string = client.DefaultRouterName
	VanillaNatsAgentName   string = client.DefaultNatsServerName
	VanillaLocalAgentName  string = "local-agent"

	// RouterMode values
	RouterModeInterior string = "interior"
	NatsModeServer     string = "server"
)
