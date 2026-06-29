package deployagent

import (
	"context"
	"fmt"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func stageRemoteWasmAirgap(ctx context.Context, namespace, host string, cfg *rsc.AgentConfiguration, ssh *rsc.SSH, pkg rsc.Package, edgelet edgeletAgent) error {
	if len(pkg.Wasm) == 0 {
		return nil
	}
	platform, err := deployairgap.ResolvePlatform(cfg.Arch)
	if err != nil {
		return fmt.Errorf("failed to resolve platform for WASM airgap: %w", err)
	}
	staged, err := deployairgap.StageAgentWasmAirgap(ctx, namespace, host, platform, ssh, pkg.Wasm)
	if err != nil {
		return fmt.Errorf("failed to transfer WASM shims: %w", err)
	}
	if len(staged) == 0 {
		return nil
	}
	return edgelet.SetWasmStaged(staged)
}
