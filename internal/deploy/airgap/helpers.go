package deployairgap

import (
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	PlatformAMD64 = "linux/amd64"
	PlatformARM64 = "linux/arm64"
)

type ContainerEngine string

const (
	EngineDocker ContainerEngine = "docker"
	EnginePodman ContainerEngine = "podman"
)

func (e ContainerEngine) Command() string {
	return string(e)
}

func ResolvePlatform(fogType *string) (string, error) {
	if fogType == nil {
		return "", util.NewInputError("Agent fog type is not configured")
	}
	value := strings.ToLower(strings.TrimSpace(*fogType))
	switch value {
	case "1", "x86", "amd64", PlatformAMD64:
		return PlatformAMD64, nil
	case "2", "arm", "arm64", PlatformARM64:
		return PlatformARM64, nil
	default:
		return "", util.NewInputError("Unsupported fog type " + *fogType)
	}
}

func ResolveContainerEngine(engine *string) (ContainerEngine, error) {
	if engine == nil {
		return "", util.NewInputError("Agent container engine configuration is missing")
	}
	value := strings.ToLower(strings.TrimSpace(*engine))
	switch value {
	case "docker":
		return EngineDocker, nil
	case "podman":
		return EnginePodman, nil
	default:
		return "", util.NewInputError("Unsupported container engine " + *engine)
	}
}

func SanitizeSegment(value string) string {
	if value == "" {
		return "default"
	}
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "value"
	}
	return result
}

// ValidateAirgapRequirements validates that required configuration is present for airgap deployment
func ValidateAirgapRequirements(agentConfig *rsc.AgentConfiguration) error {
	if agentConfig == nil {
		return util.NewInputError("Agent configuration is required for airgap deployment")
	}

	// Validate FogType
	if agentConfig.FogType == nil || *agentConfig.FogType == "" {
		return util.NewInputError("FogType is required for airgap deployment. Please specify the agent architecture (x86 or arm)")
	}

	// Validate ContainerEngine
	if agentConfig.AgentConfiguration.ContainerEngine == nil || *agentConfig.AgentConfiguration.ContainerEngine == "" {
		return util.NewInputError("ContainerEngine is required for airgap deployment. Please specify the container engine (docker or podman)")
	}

	return nil
}

// ValidateControlPlaneAirgapRequirements validates that each controller has system agent config with
// agent type (FogType) and container engine when airgap is enabled. Router and debugger are transferred
// only in the system agent phase, so system agent config is required to resolve platform.
func ValidateControlPlaneAirgapRequirements(controlPlane *rsc.RemoteControlPlane) error {
	if controlPlane == nil {
		return util.NewInputError("Control plane is required for airgap deployment")
	}
	for _, ctrl := range controlPlane.Controllers {
		if ctrl.SystemAgent == nil || ctrl.SystemAgent.AgentConfiguration == nil {
			return util.NewInputError("System agent configuration is required for airgap control plane deployment. Please specify systemAgent with agent type (x86 or arm) and container engine (docker or podman) for controller " + ctrl.Name)
		}
		if err := ValidateAirgapRequirements(ctrl.SystemAgent.AgentConfiguration); err != nil {
			return err
		}
	}
	return nil
}
