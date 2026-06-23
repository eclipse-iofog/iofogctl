package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func strPtr(v string) *string       { return &v }
func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }
func boolPtr(v bool) *bool          { return &v }

func TestEdgeletPaths(t *testing.T) {
	linux := EdgeletPaths("linux")
	if linux.ConfigFile != "/etc/edgelet/config.yaml" || linux.CertFile != "/etc/edgelet/cert.crt" {
		t.Fatalf("unexpected linux paths: %+v", linux)
	}

	win := EdgeletPaths("windows")
	if !strings.Contains(win.ConfigFile, `Edgelet\config\config.yaml`) {
		t.Fatalf("unexpected windows config path: %q", win.ConfigFile)
	}
	if !strings.Contains(win.CertFile, `Edgelet\config\cert.crt`) {
		t.Fatalf("unexpected windows cert path: %q", win.CertFile)
	}
}

func TestResolveContainerEngineURLDefaults(t *testing.T) {
	tests := []struct {
		engine string
		want   string
	}{
		{"edgelet", "unix:///run/edgelet/containerd.sock"},
		{"docker", "unix:///var/run/docker.sock"},
		{"podman", "unix:///run/podman/podman.sock"},
	}
	for _, tt := range tests {
		cfg := &client.AgentConfiguration{ContainerEngine: strPtr(tt.engine)}
		got := ResolveContainerEngineURL(tt.engine, cfg)
		if got != tt.want {
			t.Fatalf("engine=%q got %q want %q", tt.engine, got, tt.want)
		}
	}
}

func TestBuildEdgeletConfigYAMLAppliesSpecConfig(t *testing.T) {
	cfg := &client.AgentConfiguration{
		ContainerEngine:        strPtr("docker"),
		DiskLimit:              int64Ptr(50),
		MemoryLimit:            int64Ptr(4096),
		CPULimit:               int64Ptr(80),
		StatusFrequency:        float64Ptr(10),
		ChangeFrequency:        float64Ptr(10),
		WatchdogEnabled:        boolPtr(false),
		LogLevel:               strPtr("info"),
		AvailableDiskThreshold: float64Ptr(90),
		TimeZone:               "UTC",
	}

	out, err := BuildEdgeletConfigYAML("linux", "amd64", 46.2, 6.14, cfg)
	if err != nil {
		t.Fatalf("BuildEdgeletConfigYAML: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"currentProfile: production",
		"containerEngine: docker",
		"arch: amd64",
		"diskConsumptionLimit: \"50\"",
		"memoryConsumptionLimit: \"4096\"",
		"processorConsumptionLimit: \"80\"",
		"changeFrequency: \"10\"",
		"statusFrequency: \"10\"",
		"logLevel: INFO",
		"controllerCert: /etc/edgelet/cert.crt",
		"gpsCoordinates: 46.2,6.14",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("config yaml missing %q:\n%s", want, text)
		}
	}
}

func TestEdgeletProvisionCommands(t *testing.T) {
	withCert, err := EdgeletProvisionCommands("https://controller.example.com", "provision-key", "base64-ca", true)
	if err != nil {
		t.Fatalf("EdgeletProvisionCommands: %v", err)
	}
	if len(withCert) != 3 {
		t.Fatalf("expected 3 commands, got %d: %+v", len(withCert), withCert)
	}
	if withCert[0].cmd != "sudo edgelet config --a https://controller.example.com/api/v3" {
		t.Fatalf("unexpected config URL command: %q", withCert[0].cmd)
	}
	if withCert[1].cmd != "sudo edgelet config cert base64-ca" {
		t.Fatalf("unexpected cert command: %q", withCert[1].cmd)
	}
	if withCert[2].cmd != "sudo edgelet provision provision-key" {
		t.Fatalf("unexpected provision command: %q", withCert[2].cmd)
	}

	withoutCert, err := EdgeletProvisionCommands("http://localhost:51121", "key-only", "", false)
	if err != nil {
		t.Fatalf("EdgeletProvisionCommands without cert: %v", err)
	}
	if len(withoutCert) != 2 {
		t.Fatalf("expected 2 commands without cert, got %d: %+v", len(withoutCert), withoutCert)
	}
	if strings.Contains(withoutCert[0].cmd, "sudo") {
		t.Fatalf("expected no sudo prefix for local commands: %q", withoutCert[0].cmd)
	}
	if withoutCert[0].cmd != "edgelet config --a http://localhost:51121/api/v3" {
		t.Fatalf("unexpected local config command: %q", withoutCert[0].cmd)
	}
}

func TestLocalRuntimeFileExists(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "config.yaml")
	exists, err := localRuntimeFileExists(missing, false)
	if err != nil {
		t.Fatalf("localRuntimeFileExists missing: %v", err)
	}
	if exists {
		t.Fatal("expected missing file")
	}

	if err := os.WriteFile(missing, []byte("cfg"), 0o640); err != nil {
		t.Fatalf("write file: %v", err)
	}
	exists, err = localRuntimeFileExists(missing, false)
	if err != nil {
		t.Fatalf("localRuntimeFileExists present: %v", err)
	}
	if !exists {
		t.Fatal("expected existing file")
	}
}

func TestLocalRuntimeFileExistsStatPermissionDenied(t *testing.T) {
	dir := t.TempDir()
	restricted := filepath.Join(dir, "restricted")
	if err := os.Mkdir(restricted, 0o000); err != nil {
		t.Fatalf("mkdir restricted: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(restricted, 0o700) })

	target := filepath.Join(restricted, "config.yaml")
	_, err := localRuntimeFileExists(target, false)
	if err == nil {
		t.Fatal("expected permission error without sudo")
	}
}

func TestMaterializeEdgeletRuntimeTo(t *testing.T) {
	dir := t.TempDir()
	paths := EdgeletPlatformPaths{
		ConfigDir:  dir,
		ConfigFile: filepath.Join(dir, "config.yaml"),
		CertFile:   filepath.Join(dir, "cert.crt"),
	}
	spec := &EdgeletRuntimeSpec{
		Arch: "arm64",
		Agent: client.AgentConfiguration{
			ContainerEngine: strPtr("edgelet"),
		},
	}
	if err := materializeEdgeletRuntimeTo(paths, "linux", "", spec, false); err != nil {
		t.Fatalf("materializeEdgeletRuntimeTo: %v", err)
	}
	configData, err := os.ReadFile(paths.ConfigFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(configData), "arch: arm64") {
		t.Fatalf("config not written as expected: %s", configData)
	}
	certData, err := os.ReadFile(paths.CertFile)
	if err != nil {
		t.Fatalf("read cert: %v", err)
	}
	if !strings.Contains(string(certData), "BEGIN CERTIFICATE") {
		t.Fatalf("sample CA not written")
	}

	before := len(certData)
	if err := materializeEdgeletRuntimeTo(paths, "linux", "", spec, false); err != nil {
		t.Fatalf("second materializeEdgeletRuntimeTo: %v", err)
	}
	after, err := os.ReadFile(paths.CertFile)
	if err != nil {
		t.Fatalf("read cert after second run: %v", err)
	}
	if len(after) != before {
		t.Fatalf("expected existing cert to be preserved")
	}
}

func TestShouldMaterializeEdgeletRuntime(t *testing.T) {
	linuxContainer := EdgeletInstallConfig{HostOS: "linux", DeploymentType: "container"}
	if !ShouldMaterializeEdgeletRuntime(linuxContainer) {
		t.Fatal("linux container deploy should materialize host config")
	}
	darwinContainer := EdgeletInstallConfig{HostOS: "darwin", DeploymentType: "container", ContainerEngine: "docker"}
	if ShouldMaterializeEdgeletRuntime(darwinContainer) {
		t.Fatal("darwin container deploy should skip host config materialization")
	}
}

func TestBuildEdgeletBootstrapConfigCommand(t *testing.T) {
	engine := "docker"
	url := "unix:///var/run/docker.sock"
	cfg := EdgeletInstallConfig{
		HostOS:          "darwin",
		DeploymentType:  "container",
		ContainerEngine: "docker",
		Runtime: &EdgeletRuntimeSpec{
			Arch: "arm64",
			Agent: client.AgentConfiguration{
				ContainerEngine:    &engine,
				ContainerEngineURL: &url,
			},
		},
	}
	cmd, err := cfg.bootstrapConfigCommand()
	if err != nil {
		t.Fatalf("bootstrapConfigCommand: %v", err)
	}
	if !strings.Contains(cmd, "'edgelet' 'config'") && !strings.Contains(cmd, "edgelet config") {
		t.Fatalf("expected edgelet config command, got %q", cmd)
	}
	if !strings.Contains(cmd, "'--ce' 'docker'") && !strings.Contains(cmd, "--ce docker") {
		t.Fatalf("expected container engine flag, got %q", cmd)
	}
	if !strings.Contains(cmd, "'--cu' 'unix:///var/run/docker.sock'") && !strings.Contains(cmd, "--cu unix:///var/run/docker.sock") {
		t.Fatalf("expected container engine URL flag, got %q", cmd)
	}
	if strings.Contains(cmd, "controllerUrl") || strings.Contains(cmd, "controllerCert") {
		t.Fatalf("bootstrap config should omit provision-time fields, got %q", cmd)
	}
	templateOnlyFlags := []string{
		"'--tz'", "'--sec'", "'--cf'", "'--sf'", "'--sd'",
		"'--egf'", "'--pf'", "'--uf'", "'--wd'", "'--dev'",
		"'--gpsc'", "'--gps'", "'--gpsf'", "'--dt'",
		"'--d'", "'--m'", "'--p'", "'--l'", "'--dl'", "'--ld'",
	}
	for _, bad := range templateOnlyFlags {
		if strings.Contains(cmd, bad) {
			t.Fatalf("bootstrap config must not apply template defaults, found %q in %q", bad, cmd)
		}
	}
	expectedFlags := []string{"'--ce' 'docker'", "'--cu' 'unix:///var/run/docker.sock'", "'--ft' 'arm64'"}
	for _, want := range expectedFlags {
		if !strings.Contains(cmd, want) {
			t.Fatalf("expected %q in bootstrap config command, got %q", want, cmd)
		}
	}
}

func TestBuildEdgeletBootstrapConfigCommandMatchesDeploySpec(t *testing.T) {
	engine := "docker"
	logLevel := "INFO"
	cfg := EdgeletInstallConfig{
		HostOS:          "darwin",
		DeploymentType:  "container",
		ContainerEngine: "docker",
		Runtime: &EdgeletRuntimeSpec{
			Arch: "arm64",
			Agent: client.AgentConfiguration{
				ContainerEngine: &engine,
				LogLevel:        &logLevel,
			},
		},
	}
	cmd, err := cfg.bootstrapConfigCommand()
	if err != nil {
		t.Fatalf("bootstrapConfigCommand: %v", err)
	}
	for _, want := range []string{
		"'--ll' 'INFO'",
		"'--ce' 'docker'",
		"'--cu' 'unix:///var/run/docker.sock'",
		"'--ft' 'arm64'",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("expected %q in bootstrap config command, got %q", want, cmd)
		}
	}
	for _, absent := range []string{"'--tz'", "'--sec'", "'--cf'", "'--gpsc'"} {
		if strings.Contains(cmd, absent) {
			t.Fatalf("bootstrap config must not include unset spec field %q, got %q", absent, cmd)
		}
	}
}

func TestBuildEdgeletBootstrapConfigCommandSkipsDefaultEdgeletEngine(t *testing.T) {
	engine := "edgelet"
	cfg := EdgeletInstallConfig{
		HostOS:         "darwin",
		DeploymentType: "container",
		Runtime: &EdgeletRuntimeSpec{
			Agent: client.AgentConfiguration{
				ContainerEngine: &engine,
			},
		},
	}
	cmd, err := cfg.bootstrapConfigCommand()
	if err != nil {
		t.Fatalf("bootstrapConfigCommand: %v", err)
	}
	if cmd != "" {
		t.Fatalf("expected empty bootstrap config for default edgelet engine, got %q", cmd)
	}
}

func TestBootstrapEnvDesktopContainer(t *testing.T) {
	engine := "docker"
	cfg := EdgeletInstallConfig{
		HostOS:          "darwin",
		DeploymentType:  "container",
		ContainerEngine: "docker",
		Runtime: &EdgeletRuntimeSpec{
			Agent: client.AgentConfiguration{
				ContainerEngine: &engine,
			},
		},
	}
	env := cfg.bootstrapEnv(true)
	if !strings.Contains(env, "EDGELET_BOOTSTRAP_CONFIG_CMD=") {
		t.Fatalf("expected bootstrap config env, got %q", env)
	}
	if !strings.Contains(env, "EDGELET_SCRIPT_STAGE_DIR=") {
		t.Fatalf("expected stage dir in bootstrap env, got %q", env)
	}
	if !strings.Contains(env, "PATH=/tmp/edgelet-scripts/bin:") {
		t.Fatalf("expected stage bin on PATH in bootstrap env, got %q", env)
	}
}
