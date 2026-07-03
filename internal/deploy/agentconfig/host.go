package deployagentconfig

import (
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

// ResolveAgentAPIHost returns the controller API host for an agent config update.
// spec.host is used unless spec.config.host is explicitly set.
func ResolveAgentAPIHost(specHost string, deployConfig *rsc.AgentConfiguration) string {
	if deployConfig != nil && deployConfig.Host != nil && strings.TrimSpace(*deployConfig.Host) != "" {
		return strings.TrimSpace(*deployConfig.Host)
	}
	return specHost
}
