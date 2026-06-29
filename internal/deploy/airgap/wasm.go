package deployairgap

import (
	"context"
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// transferWasmBinariesHook is set by tests to mock remote WASM SCP.
var transferWasmBinariesHook func(host string, ssh *rsc.SSH, staged []wasm.StagedBinary) error

// SetTransferWasmBinariesHookForTest replaces TransferWasmBinaries with a test hook.
func SetTransferWasmBinariesHookForTest(hook func(host string, ssh *rsc.SSH, staged []wasm.StagedBinary) error) {
	transferWasmBinariesHook = hook
}

// ResetTransferWasmBinariesHookForTest clears the TransferWasmBinaries test hook.
func ResetTransferWasmBinariesHookForTest() {
	transferWasmBinariesHook = nil
}

// WasmRemoteStagingDir returns the remote directory for airgap WASM shim binaries.
func WasmRemoteStagingDir(host string) string {
	return util.JoinAgentPath(remoteAirgapDir, "wasm", SanitizeSegment(host))
}

// EnsureWasmArtifacts resolves configured WASM handlers on the operator machine.
func EnsureWasmArtifacts(ctx context.Context, namespace, platform string, wasmMap map[string]rsc.WasmPack, airgap bool) ([]wasm.StagedBinary, error) {
	if len(wasmMap) == 0 {
		return nil, nil
	}
	return wasm.ResolveWasmArtifacts(ctx, namespace, platform, toInstallWasmPacks(wasmMap), airgap)
}

// TransferWasmBinaries SCPs raw shim binaries to the remote airgap staging directory.
func TransferWasmBinaries(host string, ssh *rsc.SSH, staged []wasm.StagedBinary) error {
	if transferWasmBinariesHook != nil {
		return transferWasmBinariesHook(host, ssh, staged)
	}
	if len(staged) == 0 {
		return nil
	}
	if host == "" {
		return util.NewInputError("host is required for WASM airgap binary transfer")
	}
	if ssh == nil || ssh.User == "" || ssh.KeyFile == "" {
		return util.NewInputError("SSH configuration is required for WASM airgap binary transfer")
	}

	client, err := util.NewSecureShellClient(ssh.User, host, ssh.KeyFile)
	if err != nil {
		return err
	}
	client.SetPort(ssh.Port)
	if err := client.Connect(); err != nil {
		return err
	}
	defer util.Log(client.Disconnect)

	hostDir := WasmRemoteStagingDir(host)
	if err := client.CreateFolder(hostDir); err != nil {
		return err
	}

	for i := range staged {
		item := &staged[i]
		if item.LocalPath == "" {
			return util.NewInputError(fmt.Sprintf("local WASM binary path is required for handler %s", item.Handler))
		}

		if err := func() error {
			file, err := util.OpenValidatedFile(item.LocalPath)
			if err != nil {
				return err
			}
			defer util.IgnoreClose(file)

			info, err := file.Stat()
			if err != nil {
				return err
			}

			filename := item.CanonicalName
			if err := client.CopyTo(file, util.AddTrailingSlash(hostDir), filename, "0700", info.Size()); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}

		item.RemotePath = util.JoinAgentPath(hostDir, item.CanonicalName)
		util.PrintInfo(fmt.Sprintf("WASM shim %s transfer to %s complete", item.Handler, host))
	}
	return nil
}

// StageAgentWasmAirgap resolves and SCPs WASM shims for a remote airgap host.
func StageAgentWasmAirgap(ctx context.Context, namespace, host, platform string, ssh *rsc.SSH, wasmMap map[string]rsc.WasmPack) ([]wasm.StagedBinary, error) {
	staged, err := EnsureWasmArtifacts(ctx, namespace, platform, wasmMap, true)
	if err != nil {
		return nil, err
	}
	if len(staged) == 0 {
		return nil, nil
	}
	if err := TransferWasmBinaries(host, ssh, staged); err != nil {
		return nil, err
	}
	return staged, nil
}
