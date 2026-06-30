package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type deployManifestWriter func(data []byte, prefix string) (path string, cleanup func(), err error)

type deployManifestApplier func(manifestPath string) error

func deployWasmRuntimeClasses(hostName string, cfg EdgeletInstallConfig, write deployManifestWriter, deploy deployManifestApplier) error {
	if !wasm.ShouldInstallWasm(cfg.wasmScope()) {
		return nil
	}

	handlers := wasm.ConfiguredHandlers(cfg.Wasm)
	if len(handlers) == 0 {
		return nil
	}

	for _, handler := range handlers {
		data, err := wasm.RuntimeClassManifest(handler)
		if err != nil {
			return err
		}

		prefix := "runtimeclass-" + handler
		path, cleanup, err := write(data, prefix)
		if err != nil {
			return fmt.Errorf("write RuntimeClass manifest for %s: %w", handler, err)
		}

		if err := deploy(path); err != nil {
			cleanup()
			return fmt.Errorf("deploy RuntimeClass %q on %s: %w", handler, hostName, err)
		}
		cleanup()

		util.PrintInfo(fmt.Sprintf("RuntimeClass %s deployed on %s", handler, hostName))
	}
	return nil
}

// DeployWasmRuntimeClasses applies edgelet RuntimeClass manifests for configured WASM handlers.
func (agent *LocalEdgelet) DeployWasmRuntimeClasses() error {
	return deployWasmRuntimeClasses(agent.name, agent.cfg, agent.WriteDeployManifest, agent.DeployFromFile)
}

// DeployWasmRuntimeClasses applies edgelet RuntimeClass manifests for configured WASM handlers.
func (agent *RemoteEdgelet) DeployWasmRuntimeClasses() error {
	return deployWasmRuntimeClasses(agent.name, agent.cfg, agent.WriteDeployManifest, agent.DeployFromFile)
}
