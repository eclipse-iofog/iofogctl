package deployagent

import (
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const localAgentConflictPort = 54321

type edgeletAgent interface {
	Bootstrap() error
	Configure(controllerEndpoint string, user install.IofogUser) (string, error)
	SetVersion(version string) error
	SetContainerImage(image string) error
	SetAirgap(binPath string) error
	CustomizeProcedures(dir string, procs *install.EdgeletProcedures) error
}

func ensureLocalAgentHost(agent *rsc.LocalAgent) error {
	cfg := agent.Config
	if cfg == nil {
		cfg = &rsc.AgentConfiguration{}
		agent.Config = cfg
	}
	if cfg.Host != nil && strings.TrimSpace(*cfg.Host) != "" {
		return nil
	}
	if strings.TrimSpace(agent.Host) != "" {
		host := strings.TrimSpace(agent.Host)
		cfg.Host = &host
		return nil
	}
	ip, err := util.DetectLocalHostIPv4()
	if err != nil {
		return util.NewInputError("LocalAgent requires spec.config.host or a detectable local IPv4 address")
	}
	cfg.Host = &ip
	agent.Host = ip
	return nil
}

func checkLocalAgentPortAvailable(isSystem bool) error {
	if isSystem {
		return nil
	}
	if util.IsTCPPortOpen("127.0.0.1", localAgentConflictPort) {
		return util.NewConflictError(
			"Cannot deploy LocalAgent: an agent is already running on this host (port 54321 in use). " +
				"If you deployed LocalControlPlane, its systemAgent occupies this host — remove it first or deploy agents on remote hosts only.",
		)
	}
	return nil
}

func applyEdgeletPackage(agent edgeletAgent, pkg rsc.Package) error {
	if pkg.Container.Image != "" {
		return agent.SetContainerImage(pkg.Container.Image)
	}
	if pkg.Version != "" {
		return agent.SetVersion(pkg.Version)
	}
	return nil
}

func customizeEdgeletProcedures(agent edgeletAgent, scripts *rsc.AgentScripts) error {
	if scripts == nil {
		return nil
	}
	procs := install.EdgeletProcedures{AgentProcedures: scripts.AgentProcedures}
	return agent.CustomizeProcedures(scripts.Directory, &procs)
}
