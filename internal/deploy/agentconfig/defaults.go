package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	DefaultMessagingPort     = 5671
	DefaultEdgeRouterPort    = 45671
	DefaultInterRouterPort   = 55671
	DefaultNatsServerPort    = 4222
	DefaultNatsClusterPort   = 6222
	DefaultNatsLeafPort      = 7422
	DefaultNatsMqttPort      = 8883
	DefaultNatsHTTPPort      = 8222
	DefaultJsStorageSize     = "10G"
	DefaultJsMemoryStoreSize = "1G"
)

// PrepareForControllerAPI applies role-specific defaults and syncs arch onto the embedded SDK config.
func PrepareForControllerAPI(cfg *rsc.AgentConfiguration) {
	if cfg == nil {
		return
	}
	if iutil.IsSystemAgent(cfg) {
		ApplySystemAgentDefaults(cfg)
	} else {
		ApplyEdgeAgentDefaults(cfg)
	}
	syncArchID(cfg)
}

// ApplySystemAgentDefaults enforces interior router and server NATS with minimum port config.
func ApplySystemAgentDefaults(cfg *rsc.AgentConfiguration) {
	interior := iofog.RouterModeInterior
	cfg.RouterMode = &interior

	if cfg.EdgeRouterPort == nil {
		cfg.EdgeRouterPort = iutil.MakeIntPtr(DefaultEdgeRouterPort)
	}
	if cfg.InterRouterPort == nil {
		cfg.InterRouterPort = iutil.MakeIntPtr(DefaultInterRouterPort)
	}
	if cfg.MessagingPort == nil {
		cfg.MessagingPort = iutil.MakeIntPtr(DefaultMessagingPort)
	}

	cfg.NatsMode = iutil.MakeStrPtr(iofog.NatsModeServer)
	if cfg.NatsServerPort == nil {
		cfg.NatsServerPort = iutil.MakeIntPtr(DefaultNatsServerPort)
	}
	if cfg.NatsClusterPort == nil {
		cfg.NatsClusterPort = iutil.MakeIntPtr(DefaultNatsClusterPort)
	}
	if cfg.NatsLeafPort == nil {
		cfg.NatsLeafPort = iutil.MakeIntPtr(DefaultNatsLeafPort)
	}
	if cfg.NatsMqttPort == nil {
		cfg.NatsMqttPort = iutil.MakeIntPtr(DefaultNatsMqttPort)
	}
	if cfg.NatsHTTPPort == nil {
		cfg.NatsHTTPPort = iutil.MakeIntPtr(DefaultNatsHTTPPort)
	}
	if cfg.JsStorageSize == nil {
		cfg.JsStorageSize = iutil.MakeStrPtr(DefaultJsStorageSize)
	}
	if cfg.JsMemoryStoreSize == nil {
		cfg.JsMemoryStoreSize = iutil.MakeStrPtr(DefaultJsMemoryStoreSize)
	}
}

// ApplyEdgeAgentDefaults enforces edge router and leaf NATS defaults aligned with Controller behavior.
func ApplyEdgeAgentDefaults(cfg *rsc.AgentConfiguration) {
	if cfg.RouterMode == nil {
		edge := string(EdgeRouter)
		cfg.RouterMode = &edge
	}
	if cfg.MessagingPort == nil {
		cfg.MessagingPort = iutil.MakeIntPtr(DefaultMessagingPort)
	}

	if cfg.NatsMode == nil {
		leaf := string(NatsLeaf)
		cfg.NatsMode = &leaf
	}
	if cfg.NatsServerPort == nil {
		cfg.NatsServerPort = iutil.MakeIntPtr(DefaultNatsServerPort)
	}
	if cfg.NatsLeafPort == nil {
		cfg.NatsLeafPort = iutil.MakeIntPtr(DefaultNatsLeafPort)
	}
	if cfg.NatsMqttPort == nil {
		cfg.NatsMqttPort = iutil.MakeIntPtr(DefaultNatsMqttPort)
	}
	if cfg.NatsHTTPPort == nil {
		cfg.NatsHTTPPort = iutil.MakeIntPtr(DefaultNatsHTTPPort)
	}
	if cfg.JsStorageSize == nil {
		cfg.JsStorageSize = iutil.MakeStrPtr(DefaultJsStorageSize)
	}
	if cfg.JsMemoryStoreSize == nil {
		cfg.JsMemoryStoreSize = iutil.MakeStrPtr(DefaultJsMemoryStoreSize)
	}
}

func syncArchID(cfg *rsc.AgentConfiguration) {
	if cfg.Arch == nil {
		return
	}
	arch, found := rsc.ArchStringToID(*cfg.Arch)
	if !found {
		arch = 0
	}
	cfg.ArchID = &arch
}

// DefaultRouterConfig returns the minimum interior router config for system agents.
func DefaultRouterConfig() client.RouterConfig {
	return client.RouterConfig{
		RouterMode:      iutil.MakeStrPtr(iofog.RouterModeInterior),
		MessagingPort:   iutil.MakeIntPtr(DefaultMessagingPort),
		EdgeRouterPort:  iutil.MakeIntPtr(DefaultEdgeRouterPort),
		InterRouterPort: iutil.MakeIntPtr(DefaultInterRouterPort),
	}
}

func validateSystemAgent(config *rsc.AgentConfiguration) error {
	name := config.Name
	if config.RouterMode != nil && *config.RouterMode != iofog.RouterModeInterior {
		return util.NewInputError(fmt.Sprintf(
			"agent config %s validation failed. System agent routerMode must be interior", name))
	}
	if config.NatsMode != nil && *config.NatsMode != iofog.NatsModeServer {
		return util.NewInputError(fmt.Sprintf(
			"agent config %s validation failed. System agent natsMode must be server", name))
	}
	return validateRouterNatsFields(config, InteriorRouter, NatsServer)
}

func validateEdgeAgent(config *rsc.AgentConfiguration) error {
	routerMode := getRouterMode(config)
	natsMode := getNatsMode(config)

	if routerMode != EdgeRouter && routerMode != InteriorRouter && routerMode != NoneRouter {
		msg := "agent config %s validation failed. RouterMode has to be one of edge, interior, none. Default is: edge"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if natsMode != NatsServer && natsMode != NatsLeaf && natsMode != NatsNone {
		msg := "agent config %s validation failed. NatsMode has to be one of leaf, server, none. Default is: leaf"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	return validateRouterNatsFields(config, routerMode, natsMode)
}

func validateRouterNatsFields(config *rsc.AgentConfiguration, routerMode RouterMode, natsMode NatsMode) error {
	if routerMode != NoneRouter && config.NetworkRouter != nil {
		msg := "agent config %s validation failed. Cannot have a network if routerMode is different from none. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}
	if routerMode == NoneRouter && config.UpstreamRouters != nil && len(*config.UpstreamRouters) > 0 {
		msg := "agent config %s validation failed. Cannot have a upstreamRouters if routerMode is none"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if routerMode != InteriorRouter && (config.EdgeRouterPort != nil || config.InterRouterPort != nil) {
		msg := "agent config %s validation failed. Cannot have an edgeRouterPort or interRouterPort if routerMode is different from interior. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}
	if natsMode != NatsServer && config.NatsClusterPort != nil {
		msg := "agent config %s validation failed. Cannot have a natsClusterPort if natsMode is different from server"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	return nil
}
