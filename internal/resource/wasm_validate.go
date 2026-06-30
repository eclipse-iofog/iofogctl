package resource

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const podmanContainerEngine = "podman"

// ValidateAgentPackage validates optional package.wasm on Agent and LocalAgent specs.
func ValidateAgentPackage(label string, pkg Package, cfg *AgentConfiguration) error {
	var engine *string
	if cfg != nil {
		engine = cfg.ContainerEngine
	}
	return validatePackageWasm(label, pkg, engine)
}

func validateSystemAgentPackage(label string, systemAgent *SystemAgentConfig) error {
	if systemAgent == nil {
		return nil
	}
	var engine *string
	if systemAgent.AgentConfiguration != nil {
		engine = systemAgent.AgentConfiguration.ContainerEngine
	}
	return validatePackageWasm(label, systemAgent.Package, engine)
}

func validatePackageWasm(label string, pkg Package, containerEngine *string) error {
	if len(pkg.Wasm) == 0 {
		return nil
	}

	engine := "edgelet"
	if containerEngine != nil && strings.TrimSpace(*containerEngine) != "" {
		engine = strings.TrimSpace(*containerEngine)
	}
	if engine == podmanContainerEngine {
		return util.NewInputError(fmt.Sprintf("%s package.wasm is not supported when containerEngine is podman", label))
	}

	for handler, entry := range pkg.Wasm {
		if !isAllowedWasmHandler(handler) {
			return util.NewInputError(fmt.Sprintf("%s package.wasm contains unknown handler %q", label, handler))
		}
		if err := validateWasmPackEntry(label, handler, entry); err != nil {
			return err
		}
	}
	return nil
}

func validateWasmPackEntry(label, handler string, entry WasmPack) error {
	hasURL := strings.TrimSpace(entry.URL) != ""
	hasPath := strings.TrimSpace(entry.Path) != ""
	switch {
	case hasURL && hasPath:
		return util.NewInputError(fmt.Sprintf("%s package.wasm.%s must specify either url or path, not both", label, handler))
	case !hasURL && !hasPath:
		return util.NewInputError(fmt.Sprintf("%s package.wasm.%s requires url or path", label, handler))
	}
	if entry.SHA256 != "" {
		if err := validateWasmSHA256(label, handler, entry.SHA256); err != nil {
			return err
		}
	}
	return nil
}

func validateWasmSHA256(label, handler, value string) error {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return util.NewInputError(fmt.Sprintf("%s package.wasm.%s.sha256 must be a valid 64-character hex digest", label, handler))
	}
	return nil
}
