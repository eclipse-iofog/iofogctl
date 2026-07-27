package deployairgap

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func strPtr(v string) *string { return &v }

func TestImageLoadCommand(t *testing.T) {
	tests := []struct {
		name     string
		opts     AirgapTransferOptions
		remote   string
		contains []string
	}{
		{
			name:     "native edgelet loads archive directly",
			opts:     AirgapTransferOptions{DeploymentType: DeploymentTypeNative, Engine: EngineEdgelet},
			remote:   "/tmp/iofogctl-airgap/host/router-linux_amd64.tar.gz",
			contains: []string{"sudo edgelet image load -f \"/tmp/iofogctl-airgap/host/router-linux_amd64.tar.gz\""},
		},
		{
			name:     "native docker uses docker load",
			opts:     AirgapTransferOptions{DeploymentType: DeploymentTypeNative, Engine: EngineDocker},
			remote:   "/tmp/archive.tar.gz",
			contains: []string{"sudo -S docker load -i /tmp/archive.tar.gz"},
		},
		{
			name:     "container podman uses podman load",
			opts:     AirgapTransferOptions{DeploymentType: DeploymentTypeContainer, Engine: EnginePodman},
			remote:   "/tmp/archive.tar.gz",
			contains: []string{"sudo -S podman load -i /tmp/archive.tar.gz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ImageLoadCommand(tt.opts, tt.remote)
			for _, want := range tt.contains {
				if !containsSubstring(got, want) {
					t.Fatalf("ImageLoadCommand() = %q, want substring %q", got, want)
				}
			}
		})
	}
}

func TestAirgapTransferOptionsFromConfig(t *testing.T) {
	cfg := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType: strPtr("native"),
		},
	}
	opts, err := AirgapTransferOptionsFromConfig(cfg)
	if err != nil {
		t.Fatalf("AirgapTransferOptionsFromConfig: %v", err)
	}
	if opts.DeploymentType != DeploymentTypeNative {
		t.Fatalf("deploymentType = %q, want native", opts.DeploymentType)
	}
	if opts.Engine != EngineEdgelet {
		t.Fatalf("engine = %q, want edgelet", opts.Engine)
	}

	cfg = &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType:  strPtr("container"),
			ContainerEngine: strPtr("docker"),
		},
	}
	opts, err = AirgapTransferOptionsFromConfig(cfg)
	if err != nil {
		t.Fatalf("AirgapTransferOptionsFromConfig container: %v", err)
	}
	if opts.Engine != EngineDocker {
		t.Fatalf("engine = %q, want docker", opts.Engine)
	}
}

func TestValidateAirgapRequirements(t *testing.T) {
	nativeEdgelet := &rsc.AgentConfiguration{
		Arch: strPtr("amd64"),
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType:  strPtr("native"),
			ContainerEngine: strPtr("edgelet"),
		},
	}
	if err := ValidateAirgapRequirements(nativeEdgelet); err != nil {
		t.Fatalf("native edgelet should be valid: %v", err)
	}

	containerMissingEngine := &rsc.AgentConfiguration{
		Arch: strPtr("amd64"),
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType: strPtr("container"),
		},
	}
	if err := ValidateAirgapRequirements(containerMissingEngine); err == nil {
		t.Fatal("expected container airgap to require docker or podman")
	}

	nativeBadEngine := &rsc.AgentConfiguration{
		Arch: strPtr("amd64"),
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType:  strPtr("native"),
			ContainerEngine: strPtr("cri-o"),
		},
	}
	if err := ValidateAirgapRequirements(nativeBadEngine); err == nil {
		t.Fatal("expected unsupported native engine to fail validation")
	}
}

func TestCollectAgentAirgapImages(t *testing.T) {
	images := &RequiredImages{
		Agent:         "ghcr.io/example/edgelet:1.0",
		RouterAMD64:   "ghcr.io/example/router:1.0",
		NatsAMD64:     "ghcr.io/example/nats:1.0",
		DebuggerAMD64: "ghcr.io/example/debugger:1.0",
	}

	nativeList, err := CollectAgentAirgapImages(images, PlatformAMD64, DeploymentTypeNative)
	if err != nil {
		t.Fatalf("CollectAgentAirgapImages native: %v", err)
	}
	for _, ref := range nativeList {
		if ref == images.Agent {
			t.Fatalf("native list should not include edgelet container image, got %v", nativeList)
		}
	}
	if len(nativeList) != 3 {
		t.Fatalf("native list len = %d, want 3 (%v)", len(nativeList), nativeList)
	}

	containerList, err := CollectAgentAirgapImages(images, PlatformAMD64, DeploymentTypeContainer)
	if err != nil {
		t.Fatalf("CollectAgentAirgapImages container: %v", err)
	}
	if containerList[0] != images.Agent {
		t.Fatalf("container list should include edgelet image first, got %v", containerList)
	}
}

func TestCollectAgentAirgapImagesDedupesDuplicateDebuggerRefs(t *testing.T) {
	sameDebugger := "ghcr.io/example/debugger:1.0"
	images := &RequiredImages{
		RouterAMD64:     "ghcr.io/example/router:1.0",
		NatsAMD64:       "ghcr.io/example/nats:1.0",
		DebuggerAMD64:   sameDebugger,
		DebuggerARM64:   sameDebugger,
		DebuggerRISCV64: sameDebugger,
		DebuggerARM:     sameDebugger,
	}

	list, err := CollectAgentAirgapImages(images, PlatformAMD64, DeploymentTypeNative)
	if err != nil {
		t.Fatalf("CollectAgentAirgapImages: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("list len = %d, want 3 (%v)", len(list), list)
	}
	debuggerCount := 0
	for _, ref := range list {
		if ref == sameDebugger {
			debuggerCount++
		}
	}
	if debuggerCount != 1 {
		t.Fatalf("debugger ref count = %d, want 1 (%v)", debuggerCount, list)
	}
}

func TestCollectAgentAirgapImagesUsesPlatformSpecificDebugger(t *testing.T) {
	images := &RequiredImages{
		RouterARM64:   "ghcr.io/example/router:arm64",
		NatsARM64:     "ghcr.io/example/nats:arm64",
		DebuggerAMD64: "ghcr.io/example/debugger:amd64",
		DebuggerARM64: "ghcr.io/example/debugger:arm64",
	}

	list, err := CollectAgentAirgapImages(images, PlatformARM64, DeploymentTypeNative)
	if err != nil {
		t.Fatalf("CollectAgentAirgapImages: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("list len = %d, want 3 (%v)", len(list), list)
	}
	if list[2] != images.DebuggerARM64 {
		t.Fatalf("debugger = %q, want %q", list[2], images.DebuggerARM64)
	}
}

func TestCollectSystemMicroserviceAirgapImages(t *testing.T) {
	images := &RequiredImages{
		RouterAMD64:   "ghcr.io/example/router:1.0",
		NatsAMD64:     "ghcr.io/example/nats:1.0",
		DebuggerAMD64: "ghcr.io/example/debugger:1.0",
	}

	list, err := CollectSystemMicroserviceAirgapImages(images, PlatformAMD64)
	if err != nil {
		t.Fatalf("CollectSystemMicroserviceAirgapImages: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("list len = %d, want 3 (%v)", len(list), list)
	}
}

func TestCollectControllerHostAirgapImagesIncludesSystemMicroservices(t *testing.T) {
	images := &RequiredImages{
		Controller:    "ghcr.io/example/controller:1.0",
		RouterAMD64:   "ghcr.io/example/router:1.0",
		NatsAMD64:     "ghcr.io/example/nats:1.0",
		DebuggerAMD64: "ghcr.io/example/debugger:1.0",
	}

	list, err := CollectControllerHostAirgapImages(images, PlatformAMD64)
	if err != nil {
		t.Fatalf("CollectControllerHostAirgapImages: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("list len = %d, want 4 (%v)", len(list), list)
	}
	if list[0] != images.Controller {
		t.Fatalf("controller = %q, want %q", list[0], images.Controller)
	}
}

func TestCollectControllerHostAirgapImagesDedupesDuplicateSystemRefs(t *testing.T) {
	same := "ghcr.io/example/shared:1.0"
	images := &RequiredImages{
		Controller:      "ghcr.io/example/controller:1.0",
		RouterAMD64:     same,
		NatsAMD64:       same,
		DebuggerAMD64:   same,
		DebuggerARM64:   same,
		DebuggerRISCV64: same,
		DebuggerARM:     same,
	}

	list, err := CollectControllerHostAirgapImages(images, PlatformAMD64)
	if err != nil {
		t.Fatalf("CollectControllerHostAirgapImages: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2 (%v)", len(list), list)
	}
}

func TestDedupeNonEmpty(t *testing.T) {
	got := dedupeNonEmpty([]string{"a", "b", "a", "", "c", "b"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("dedupeNonEmpty = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dedupeNonEmpty = %v, want %v", got, want)
		}
	}
}

func TestValidateControllerAirgapRequirementsRejectsMissingConfig(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-2"}
	if err := ValidateControllerAirgapRequirements(ctrl); err == nil {
		t.Fatal("expected error when systemAgent.config is missing")
	}
}

func TestControllerAirgapEnabled(t *testing.T) {
	cp := &rsc.RemoteControlPlane{Airgap: true}
	ctrl := &rsc.RemoteController{Airgap: false}
	if !ControllerAirgapEnabled(cp, ctrl) {
		t.Fatal("expected airgap when control plane airgap is true")
	}

	cp = &rsc.RemoteControlPlane{Airgap: false}
	ctrl = &rsc.RemoteController{Airgap: true}
	if !ControllerAirgapEnabled(cp, ctrl) {
		t.Fatal("expected airgap when controller airgap is true")
	}

	cp = &rsc.RemoteControlPlane{Airgap: false}
	ctrl = &rsc.RemoteController{Airgap: false}
	if ControllerAirgapEnabled(cp, ctrl) {
		t.Fatal("expected no airgap when both flags are false")
	}
}

func TestControllerAirgapLoadOptions(t *testing.T) {
	cfg := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			ContainerEngine: strPtr("docker"),
		},
	}
	opts, err := ControllerAirgapLoadOptions(cfg)
	if err != nil {
		t.Fatalf("ControllerAirgapLoadOptions: %v", err)
	}
	if opts.Engine != EngineDocker || opts.DeploymentType != DeploymentTypeContainer {
		t.Fatalf("unexpected opts: %+v", opts)
	}

	cfg = &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			ContainerEngine: strPtr("edgelet"),
		},
	}
	if _, err := ControllerAirgapLoadOptions(cfg); err == nil {
		t.Fatal("expected controller load options to reject edgelet engine")
	}
}

func TestPlatformToOSArch(t *testing.T) {
	osName, archName, err := PlatformToOSArch(PlatformARM64)
	if err != nil {
		t.Fatalf("PlatformToOSArch: %v", err)
	}
	if osName != "linux" || archName != "arm64" {
		t.Fatalf("got %s/%s, want linux/arm64", osName, archName)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && strings.Contains(s, sub)
}

func TestResolveAgentDeploymentDefaultsNative(t *testing.T) {
	cfg := &rsc.AgentConfiguration{}
	if ResolveAgentDeployment(cfg, "") {
		t.Fatal("expected native deployment")
	}
	if cfg.DeploymentType == nil || *cfg.DeploymentType != DeploymentTypeNative {
		t.Fatalf("deploymentType = %v, want native", cfg.DeploymentType)
	}
	if cfg.ContainerEngine == nil || *cfg.ContainerEngine != string(EngineEdgelet) {
		t.Fatalf("containerEngine = %v, want edgelet", cfg.ContainerEngine)
	}
}

func TestResolveAgentDeploymentContainerImage(t *testing.T) {
	cfg := &rsc.AgentConfiguration{}
	if !ResolveAgentDeployment(cfg, "ghcr.io/example/edgelet:1.0") {
		t.Fatal("expected container deployment when image is set")
	}
	if *cfg.DeploymentType != DeploymentTypeContainer {
		t.Fatalf("deploymentType = %q, want container", *cfg.DeploymentType)
	}
}

func TestEdgeletInstallConfigDefaults(t *testing.T) {
	cfg := EdgeletInstallConfig("linux", &rsc.AgentConfiguration{}, rsc.Package{})
	if cfg.DeploymentType != DeploymentTypeNative {
		t.Fatalf("deploymentType = %q, want native", cfg.DeploymentType)
	}
	if cfg.ContainerEngine != string(EngineEdgelet) {
		t.Fatalf("containerEngine = %q, want edgelet", cfg.ContainerEngine)
	}
	if cfg.Runtime == nil {
		t.Fatal("expected runtime spec")
	}
}
