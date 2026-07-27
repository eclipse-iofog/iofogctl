package deployairgap

import (
	"fmt"
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	PlatformAMD64   = "linux/amd64"
	PlatformARM64   = "linux/arm64"
	PlatformRISCV64 = "linux/riscv64"
	PlatformARM     = "linux/arm"

	DeploymentTypeNative    = "native"
	DeploymentTypeContainer = "container"
)

type ContainerEngine string

const (
	EngineEdgelet ContainerEngine = "edgelet"
	EngineDocker  ContainerEngine = "docker"
	EnginePodman  ContainerEngine = "podman"
)

func (e ContainerEngine) Command() string {
	return string(e)
}

// AirgapTransferOptions selects image load commands for an airgap transfer.
type AirgapTransferOptions struct {
	DeploymentType string
	Engine         ContainerEngine
}

func ResolvePlatform(arch *string) (string, error) {
	if arch == nil {
		return "", util.NewInputError("Agent fog type is not configured")
	}
	value := strings.ToLower(strings.TrimSpace(*arch))
	switch value {
	case "1", "x86", "amd64", PlatformAMD64:
		return PlatformAMD64, nil
	case "2", "arm", "arm64", PlatformARM64:
		return PlatformARM64, nil
	case "3", "riscv64", PlatformRISCV64:
		return PlatformRISCV64, nil
	default:
		return "", util.NewInputError("Unsupported fog type " + *arch)
	}
}

func ResolveDeploymentType(deploymentType *string) string {
	if deploymentType == nil {
		return DeploymentTypeNative
	}
	value := strings.ToLower(strings.TrimSpace(*deploymentType))
	switch value {
	case DeploymentTypeContainer:
		return DeploymentTypeContainer
	default:
		return DeploymentTypeNative
	}
}

func IsNativeDeployment(deploymentType string) bool {
	return ResolveDeploymentType(&deploymentType) == DeploymentTypeNative
}

func ResolveContainerEngine(engine *string) (ContainerEngine, error) {
	if engine == nil {
		return "", util.NewInputError("Agent container engine configuration is missing")
	}
	value := strings.ToLower(strings.TrimSpace(*engine))
	switch value {
	case "edgelet":
		return EngineEdgelet, nil
	case "docker":
		return EngineDocker, nil
	case "podman":
		return EnginePodman, nil
	default:
		return "", util.NewInputError("Unsupported container engine " + *engine)
	}
}

// AirgapTransferOptionsFromConfig resolves deployment type and engine defaults for airgap transfers.
func AirgapTransferOptionsFromConfig(cfg *rsc.AgentConfiguration) (AirgapTransferOptions, error) {
	if cfg == nil {
		return AirgapTransferOptions{}, util.NewInputError("Agent configuration is required for airgap deployment")
	}
	deploymentType := ResolveDeploymentType(cfg.DeploymentType)
	engineStr := ""
	if cfg.ContainerEngine != nil {
		engineStr = strings.ToLower(strings.TrimSpace(*cfg.ContainerEngine))
	}
	if engineStr == "" {
		if deploymentType == DeploymentTypeNative {
			return AirgapTransferOptions{DeploymentType: deploymentType, Engine: EngineEdgelet}, nil
		}
		return AirgapTransferOptions{}, util.NewInputError("ContainerEngine is required for container airgap deployment")
	}
	engine, err := ResolveContainerEngine(&engineStr)
	if err != nil {
		return AirgapTransferOptions{}, err
	}
	return AirgapTransferOptions{DeploymentType: deploymentType, Engine: engine}, nil
}

// PlatformToOSArch splits a platform string such as linux/amd64 into OS and arch parts.
func PlatformToOSArch(platform string) (osName, archName string, err error) {
	parts := strings.Split(platform, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", util.NewInternalError("invalid platform specification " + platform)
	}
	return parts[0], parts[1], nil
}

// ImageLoadCommand builds the remote shell command used to import a transferred image archive.
func ImageLoadCommand(opts AirgapTransferOptions, remoteArchivePath string) string {
	if IsNativeDeployment(opts.DeploymentType) && opts.Engine == EngineEdgelet {
		return fmt.Sprintf("sudo edgelet image load -f %q", remoteArchivePath)
	}
	return fmt.Sprintf("sudo -S %s load -i %s", opts.Engine.Command(), remoteArchivePath)
}

// CollectSystemMicroserviceAirgapImages returns router, NATS, and debugger refs for one platform.
func CollectSystemMicroserviceAirgapImages(images *RequiredImages, platform string) ([]string, error) {
	if images == nil {
		return nil, util.NewInternalError("required images are missing")
	}

	imageList := make([]string, 0, 3)

	routerImage, err := GetImageForPlatform(images, platform)
	if err != nil {
		return nil, err
	}
	if routerImage != "" {
		imageList = append(imageList, routerImage)
	}

	natsImage, err := GetNatsForPlatform(images, platform)
	if err != nil {
		return nil, err
	}
	if natsImage != "" {
		imageList = append(imageList, natsImage)
	}

	debuggerImage, err := GetDebuggerForPlatform(images, platform)
	if err != nil {
		return nil, err
	}
	if debuggerImage != "" {
		imageList = append(imageList, debuggerImage)
	}

	return dedupeNonEmpty(imageList), nil
}

// CollectAgentAirgapImages returns image refs to transfer for an agent airgap deploy.
// Native deployments skip the edgelet container image because the raw binary is transferred separately.
func CollectAgentAirgapImages(images *RequiredImages, platform, deploymentType string) ([]string, error) {
	if images == nil {
		return nil, util.NewInternalError("required images are missing")
	}

	imageList := make([]string, 0, 4)
	if !IsNativeDeployment(deploymentType) && images.Agent != "" {
		imageList = append(imageList, images.Agent)
	}

	sysImages, err := CollectSystemMicroserviceAirgapImages(images, platform)
	if err != nil {
		return nil, err
	}
	imageList = append(imageList, sysImages...)

	return dedupeNonEmpty(imageList), nil
}

// CollectControllerHostAirgapImages returns image refs to transfer for a controller systemAgent host.
func CollectControllerHostAirgapImages(images *RequiredImages, platform string) ([]string, error) {
	if images == nil {
		return nil, util.NewInternalError("required images are missing")
	}

	imageList := make([]string, 0, 4)
	if images.Controller != "" {
		imageList = append(imageList, images.Controller)
	}

	sysImages, err := CollectSystemMicroserviceAirgapImages(images, platform)
	if err != nil {
		return nil, err
	}
	imageList = append(imageList, sysImages...)

	return dedupeNonEmpty(imageList), nil
}

func dedupeNonEmpty(refs []string) []string {
	if len(refs) == 0 {
		return refs
	}
	seen := make(map[string]struct{}, len(refs))
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

// ControllerAirgapLoadOptions returns docker/podman load options for controller container images.
func ControllerAirgapLoadOptions(cfg *rsc.AgentConfiguration) (AirgapTransferOptions, error) {
	if cfg == nil || cfg.ContainerEngine == nil {
		return AirgapTransferOptions{}, util.NewInputError("containerEngine docker or podman is required to load controller images in airgap mode")
	}
	engineStr := strings.ToLower(strings.TrimSpace(*cfg.ContainerEngine))
	switch engineStr {
	case "docker", "podman":
		engine, err := ResolveContainerEngine(&engineStr)
		if err != nil {
			return AirgapTransferOptions{}, err
		}
		return AirgapTransferOptions{DeploymentType: DeploymentTypeContainer, Engine: engine}, nil
	default:
		return AirgapTransferOptions{}, util.NewInputError("controller airgap image load requires containerEngine docker or podman on the controller host")
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

func syncArchFromID(cfg *rsc.AgentConfiguration) {
	if cfg == nil {
		return
	}
	if cfg.Arch != nil && strings.TrimSpace(*cfg.Arch) != "" {
		return
	}
	if cfg.ArchID == nil {
		return
	}
	name, ok := rsc.ArchIDToString(int(*cfg.ArchID))
	if !ok || name == "" || name == "auto" {
		return
	}
	cfg.Arch = iutil.MakeStrPtr(name)
}

// NormalizeAgentArch ensures cfg.Arch is set, deriving it from cfg.ArchID when needed.
func NormalizeAgentArch(cfg *rsc.AgentConfiguration) error {
	if cfg == nil {
		return util.NewInputError("Agent configuration is required for airgap deployment")
	}
	syncArchFromID(cfg)
	if cfg.Arch == nil || strings.TrimSpace(*cfg.Arch) == "" {
		if cfg.ArchID != nil {
			return util.NewInputError(fmt.Sprintf("Unsupported archId %d for airgap deployment", *cfg.ArchID))
		}
		return util.NewInputError("Arch or archId is required for airgap deployment. Please specify the agent architecture (amd64, arm64, riscv64, arm)")
	}
	return nil
}

// ValidateAirgapRequirements validates that required configuration is present for airgap deployment.
func ValidateAirgapRequirements(agentConfig *rsc.AgentConfiguration) error {
	if agentConfig == nil {
		return util.NewInputError("Agent configuration is required for airgap deployment")
	}

	if err := NormalizeAgentArch(agentConfig); err != nil {
		return err
	}

	deploymentType := ResolveDeploymentType(agentConfig.DeploymentType)
	engineStr := ""
	if agentConfig.ContainerEngine != nil {
		engineStr = strings.ToLower(strings.TrimSpace(*agentConfig.ContainerEngine))
	}

	if deploymentType == DeploymentTypeContainer {
		if engineStr != "docker" && engineStr != "podman" {
			return util.NewInputError("ContainerEngine docker or podman is required for container airgap deployment")
		}
		return nil
	}

	if engineStr == "" || engineStr == "edgelet" {
		return nil
	}
	if engineStr == "docker" || engineStr == "podman" {
		return nil
	}
	return util.NewInputError("Unsupported container engine " + engineStr + " for native airgap deployment")
}

// ControllerAirgapEnabled reports whether airgap transfer should run for a controller host.
func ControllerAirgapEnabled(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) bool {
	if cp != nil && cp.Airgap {
		return true
	}
	if ctrl != nil && ctrl.Airgap {
		return true
	}
	return false
}

// ValidateControllerAirgapRequirements validates system agent config for a single airgap controller.
func ValidateControllerAirgapRequirements(ctrl *rsc.RemoteController) error {
	if ctrl == nil {
		return util.NewInputError("Controller is required for airgap deployment")
	}
	if ctrl.SystemAgent == nil || ctrl.SystemAgent.AgentConfiguration == nil {
		return util.NewInputError("System agent configuration is required for airgap controller deployment. Please specify systemAgent with agent type (x86 or arm) and container engine for controller " + ctrl.Name)
	}
	return ValidateAirgapRequirements(ctrl.SystemAgent.AgentConfiguration)
}

// ValidateControlPlaneAirgapRequirements validates controllers that require airgap deployment.
// Router, NATS, and debugger are transferred on the controller host during DeployHostEdgelet,
// before the control plane manifest is applied.
func ValidateControlPlaneAirgapRequirements(controlPlane *rsc.RemoteControlPlane) error {
	if controlPlane == nil {
		return util.NewInputError("Control plane is required for airgap deployment")
	}
	for idx := range controlPlane.Controllers {
		ctrl := &controlPlane.Controllers[idx]
		if !ControllerAirgapEnabled(controlPlane, ctrl) {
			continue
		}
		if err := ValidateControllerAirgapRequirements(ctrl); err != nil {
			return err
		}
	}
	return nil
}
