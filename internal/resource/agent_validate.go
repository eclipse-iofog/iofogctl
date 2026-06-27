package resource

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func validateSystemAgentRouterNats(label string, cfg *AgentConfiguration) error {
	if cfg == nil {
		return nil
	}
	if cfg.RouterMode != nil && *cfg.RouterMode != iofog.RouterModeInterior {
		return util.NewInputError(fmt.Sprintf("%s systemAgent.config.routerMode must be %q when set", label, iofog.RouterModeInterior))
	}
	if cfg.NatsMode != nil && *cfg.NatsMode != iofog.NatsModeServer {
		return util.NewInputError(fmt.Sprintf("%s systemAgent.config.natsMode must be %q when set", label, iofog.NatsModeServer))
	}
	return nil
}
