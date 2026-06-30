package deployairgap

import (
	"runtime"
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install/wasm"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func toInstallWasmPacks(wasmMap map[string]rsc.WasmPack) map[string]wasm.Pack {
	if len(wasmMap) == 0 {
		return nil
	}
	out := make(map[string]wasm.Pack, len(wasmMap))
	for handler, pack := range wasmMap {
		out[handler] = wasm.Pack{
			URL:    pack.URL,
			Path:   pack.Path,
			SHA256: pack.SHA256,
		}
	}
	return out
}

// EnsureAgentConfig returns a non-nil agent configuration struct.
func EnsureAgentConfig(agent *rsc.AgentConfiguration) *rsc.AgentConfiguration {
	if agent == nil {
		return &rsc.AgentConfiguration{}
	}
	return agent
}

// ResolveAgentDeployment normalizes deployment defaults and reports whether container install is selected.
func ResolveAgentDeployment(cfg *rsc.AgentConfiguration, containerImage string) bool {
	cfg = EnsureAgentConfig(cfg)

	useContainer := cfg.DeploymentType != nil && strings.EqualFold(strings.TrimSpace(*cfg.DeploymentType), DeploymentTypeContainer)
	if containerImage != "" {
		useContainer = true
	}

	if useContainer {
		cfg.DeploymentType = iutil.MakeStrPtr(DeploymentTypeContainer)
	} else {
		cfg.DeploymentType = iutil.MakeStrPtr(DeploymentTypeNative)
	}

	if cfg.ContainerEngine == nil || strings.TrimSpace(*cfg.ContainerEngine) == "" {
		cfg.ContainerEngine = iutil.MakeStrPtr(string(EngineEdgelet))
	}

	return useContainer
}

func edgeletRuntimeSpec(cfg *rsc.AgentConfiguration) *install.EdgeletRuntimeSpec {
	if cfg == nil {
		return nil
	}
	arch := "auto"
	if cfg.Arch != nil && strings.TrimSpace(*cfg.Arch) != "" {
		arch = strings.TrimSpace(*cfg.Arch)
	}
	return &install.EdgeletRuntimeSpec{
		Arch:      arch,
		Latitude:  cfg.Latitude,
		Longitude: cfg.Longitude,
		Agent:     cfg.AgentConfiguration,
	}
}

// EdgeletInstallConfig builds layered install settings from agent YAML spec fields.
func EdgeletInstallConfig(hostOS string, cfg *rsc.AgentConfiguration, pkg rsc.Package) install.EdgeletInstallConfig {
	cfg = EnsureAgentConfig(cfg)
	ResolveAgentDeployment(cfg, pkg.Container.Image)

	arch := "auto"
	if cfg.Arch != nil && strings.TrimSpace(*cfg.Arch) != "" {
		arch = strings.TrimSpace(*cfg.Arch)
	}

	engine := string(EngineEdgelet)
	if cfg.ContainerEngine != nil && strings.TrimSpace(*cfg.ContainerEngine) != "" {
		engine = strings.TrimSpace(*cfg.ContainerEngine)
	}

	deploymentType := ResolveDeploymentType(cfg.DeploymentType)

	installCfg := install.EdgeletInstallConfig{
		HostOS:          hostOS,
		Arch:            arch,
		ContainerEngine: engine,
		DeploymentType:  deploymentType,
		ContainerImage:  pkg.Container.Image,
		TimeZone:        cfg.TimeZone,
		Runtime:         edgeletRuntimeSpec(cfg),
		Wasm:            toInstallWasmPacks(pkg.Wasm),
	}
	if pkg.Version != "" {
		installCfg.Version = pkg.Version
	}
	return installCfg
}

// LocalEdgeletHostOS returns the normalized edgelet OS name for the local runtime.
func LocalEdgeletHostOS() string {
	osName, err := util.NormalizeEdgeletOS(runtime.GOOS)
	if err != nil {
		return "linux"
	}
	return osName
}
