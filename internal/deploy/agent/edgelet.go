package deployagent

import (
	"context"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	deployvalidate "github.com/eclipse-iofog/iofogctl/internal/deploy/validate"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type edgeletAgent interface {
	Bootstrap() error
	PrepareWasm(ctx context.Context, namespace string) error
	SetWasmStaged(staged []wasm.StagedBinary) error
	Configure(controllerEndpoint string, user install.IofogUser, sdkOpt client.Options) (string, error)
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
	return deployvalidate.LocalAgentPortAvailable(isSystem)
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
