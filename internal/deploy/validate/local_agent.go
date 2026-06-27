package validate

import "github.com/eclipse-iofog/iofogctl/pkg/util"

const LocalAgentConflictPort = 54321

func LocalAgentPortAvailable(isSystem bool) error {
	if isSystem {
		return nil
	}
	if util.IsTCPPortOpen("127.0.0.1", LocalAgentConflictPort) {
		return util.NewConflictError(
			"Cannot deploy LocalAgent: an agent is already running on this host (port 54321 in use). " +
				"If you deployed LocalControlPlane, its systemAgent occupies this host — remove it first or deploy agents on remote hosts only.",
		)
	}
	return nil
}
