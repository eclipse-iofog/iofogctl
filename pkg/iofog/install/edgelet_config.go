package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	edgeletConfigAsset    = "edgelet/edgelet-config.yaml"
	edgeletSampleCAAsset  = "edgelet/edgelet-controller-ca.crt"
	edgeletDefaultProfile = "production"
)

// EdgeletRuntimeSpec carries spec.config fields used to materialize edgelet profile YAML.
type EdgeletRuntimeSpec struct {
	Arch      string
	Latitude  float64
	Longitude float64
	Agent     client.AgentConfiguration
}

// EdgeletPlatformPaths holds on-host edgelet config and cert locations.
type EdgeletPlatformPaths struct {
	ConfigDir  string
	ConfigFile string
	CertFile   string
}

// EdgeletPaths returns config and cert paths for a host OS (linux, darwin, windows).
func EdgeletPaths(hostOS string) EdgeletPlatformPaths {
	switch normalizeEdgeletHostOS(hostOS) {
	case "windows":
		base := `%ProgramData%\Edgelet\config`
		return EdgeletPlatformPaths{
			ConfigDir:  base,
			ConfigFile: base + `\config.yaml`,
			CertFile:   base + `\cert.crt`,
		}
	default:
		return EdgeletPlatformPaths{
			ConfigDir:  "/etc/edgelet",
			ConfigFile: "/etc/edgelet/config.yaml",
			CertFile:   "/etc/edgelet/cert.crt",
		}
	}
}

func normalizeEdgeletHostOS(hostOS string) string {
	switch strings.ToLower(strings.TrimSpace(hostOS)) {
	case "darwin", "macos", "osx":
		return "darwin"
	case "windows", "windows_nt":
		return "windows"
	default:
		return "linux"
	}
}

func loadEdgeletConfigTemplate() (map[string]interface{}, error) {
	raw, err := util.GetStaticFile(edgeletConfigAsset)
	if err != nil {
		return nil, err
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("parse edgelet config template: %w", err)
	}
	return doc, nil
}

func loadEdgeletSampleCA() ([]byte, error) {
	raw, err := util.GetStaticFile(edgeletSampleCAAsset)
	if err != nil {
		return nil, err
	}
	return []byte(raw), nil
}

func productionProfile(doc map[string]interface{}) (map[string]interface{}, error) {
	profiles, ok := doc["profiles"].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("edgelet config template missing profiles")
	}
	raw, ok := profiles[edgeletDefaultProfile]
	if !ok {
		return nil, fmt.Errorf("edgelet config template missing %q profile", edgeletDefaultProfile)
	}
	profile, ok := raw.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("edgelet config template profile has unexpected shape")
	}
	out := make(map[string]interface{}, len(profile))
	for k, v := range profile {
		key, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("edgelet config template profile key has unexpected type")
		}
		out[key] = v
	}
	return out, nil
}

func defaultContainerEngineURL(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case "docker":
		return "unix:///var/run/docker.sock"
	case "podman":
		return "unix:///run/podman/podman.sock"
	default:
		return "unix:///run/edgelet/containerd.sock"
	}
}

func resolveContainerEngine(cfg *client.AgentConfiguration) string {
	if cfg != nil && cfg.ContainerEngine != nil && strings.TrimSpace(*cfg.ContainerEngine) != "" {
		return strings.TrimSpace(*cfg.ContainerEngine)
	}
	return "edgelet"
}

func resolveContainerEngineURL(engine string, cfg *client.AgentConfiguration) string {
	if cfg != nil && cfg.ContainerEngineURL != nil && strings.TrimSpace(*cfg.ContainerEngineURL) != "" {
		return strings.TrimSpace(*cfg.ContainerEngineURL)
	}
	return defaultContainerEngineURL(engine)
}

func boolToOnOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func normalizeLogLevel(level string) string {
	level = strings.TrimSpace(level)
	if level == "" {
		return "INFO"
	}
	return strings.ToUpper(level)
}

func applyEdgeletProfileOverrides(profile map[string]interface{}, hostOS, arch string, latitude, longitude float64, cfg *client.AgentConfiguration) {
	paths := EdgeletPaths(hostOS)
	profile["controllerCert"] = paths.CertFile

	if arch != "" {
		profile["arch"] = arch
	}

	engine := resolveContainerEngine(cfg)
	profile["containerEngine"] = engine
	profile["containerEngineUrl"] = resolveContainerEngineURL(engine, cfg)

	if cfg == nil {
		return
	}

	if cfg.NetworkInterface != nil && *cfg.NetworkInterface != "" {
		profile["networkInterface"] = *cfg.NetworkInterface
	}
	if cfg.DiskLimit != nil {
		profile["diskConsumptionLimit"] = formatInt(*cfg.DiskLimit)
	}
	if cfg.DiskDirectory != nil && *cfg.DiskDirectory != "" {
		profile["diskDirectory"] = *cfg.DiskDirectory
	}
	if cfg.MemoryLimit != nil {
		profile["memoryConsumptionLimit"] = formatInt(*cfg.MemoryLimit)
	}
	if cfg.CPULimit != nil {
		profile["processorConsumptionLimit"] = formatInt(*cfg.CPULimit)
	}
	if cfg.LogLimit != nil {
		profile["logDiskConsumptionLimit"] = formatInt(*cfg.LogLimit)
	}
	if cfg.LogDirectory != nil && *cfg.LogDirectory != "" {
		profile["logDiskDirectory"] = *cfg.LogDirectory
	}
	if cfg.LogFileCount != nil {
		profile["logFileCount"] = formatInt(*cfg.LogFileCount)
	}
	if cfg.StatusFrequency != nil {
		profile["statusFrequency"] = formatFloat(*cfg.StatusFrequency)
	}
	if cfg.ChangeFrequency != nil {
		profile["changeFrequency"] = formatFloat(*cfg.ChangeFrequency)
	}
	if cfg.DeviceScanFrequency != nil {
		profile["scanDevicesFreq"] = formatFloat(*cfg.DeviceScanFrequency)
	}
	if cfg.WatchdogEnabled != nil {
		profile["watchdogEnabled"] = boolToOnOff(*cfg.WatchdogEnabled)
	}
	if cfg.GpsMode != nil && *cfg.GpsMode != "" {
		profile["gps"] = *cfg.GpsMode
	}
	if latitude != 0 || longitude != 0 {
		profile["gpsCoordinates"] = fmt.Sprintf("%g,%g", latitude, longitude)
	}
	if cfg.GpsDevice != nil {
		profile["gpsDevice"] = *cfg.GpsDevice
	}
	if cfg.GpsScanFrequency != nil {
		profile["gpsScanFrequency"] = formatFloat(*cfg.GpsScanFrequency)
	}
	if cfg.EdgeGuardFrequency != nil {
		profile["edgeGuardFreq"] = formatFloat(*cfg.EdgeGuardFrequency)
	}
	if cfg.PruningFrequency != nil {
		profile["pruningFrequency"] = formatFloat(*cfg.PruningFrequency)
	}
	if cfg.AvailableDiskThreshold != nil {
		profile["availableDiskThreshold"] = formatFloat(*cfg.AvailableDiskThreshold)
	}
	if cfg.LogLevel != nil && *cfg.LogLevel != "" {
		profile["logLevel"] = normalizeLogLevel(*cfg.LogLevel)
	}
	if cfg.TimeZone != "" {
		profile["timeZone"] = cfg.TimeZone
	}
}

// BuildEdgeletConfigYAML renders edgelet runtime config from the embedded template and spec.config overrides.
func BuildEdgeletConfigYAML(hostOS, arch string, latitude, longitude float64, cfg *client.AgentConfiguration) ([]byte, error) {
	doc, err := loadEdgeletConfigTemplate()
	if err != nil {
		return nil, err
	}
	profile, err := productionProfile(doc)
	if err != nil {
		return nil, err
	}
	applyEdgeletProfileOverrides(profile, hostOS, arch, latitude, longitude, cfg)

	profiles := map[interface{}]interface{}{
		edgeletDefaultProfile: profile,
	}
	doc["currentProfile"] = edgeletDefaultProfile
	doc["profiles"] = profiles

	out, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal edgelet config: %w", err)
	}
	return out, nil
}

// IsDesktopContainerHost reports macOS/Windows hosts that use desktop container runtimes.
func IsDesktopContainerHost(hostOS string) bool {
	switch normalizeEdgeletHostOS(hostOS) {
	case "darwin", "windows":
		return true
	default:
		return false
	}
}

// IsDesktopContainerDeploy reports container edgelet on a desktop-class host OS.
func IsDesktopContainerDeploy(cfg EdgeletInstallConfig) bool {
	return !cfg.native() && IsDesktopContainerHost(cfg.hostOS())
}

// ShouldMaterializeEdgeletRuntime reports whether host-side config files should be written before start.
func ShouldMaterializeEdgeletRuntime(cfg EdgeletInstallConfig) bool {
	return !IsDesktopContainerDeploy(cfg)
}

var edgeletBootstrapConfigSkipKeys = map[string]struct{}{
	"controllerUrl":  {},
	"iofogUuid":      {},
	"controllerCert": {},
}

// edgeletProfileKeyToConfigFlag maps edgelet-config.yaml profile keys to edgelet config CLI
// short aliases (see edgelet/docs/cli/generated/edgelet_config.md).
var edgeletProfileKeyToConfigFlag = map[string]string{
	"arch":                      "--ft",
	"availableDiskThreshold":    "--dt",
	"changeFrequency":           "--cf",
	"containerEngine":           "--ce",
	"containerEngineUrl":        "--cu",
	"devMode":                   "--dev",
	"diskDirectory":             "--dl",
	"diskConsumptionLimit":      "--d",
	"edgeGuardFreq":             "--egf",
	"gps":                       "--gps",
	"gpsCoordinates":            "--gpsc",
	"gpsDevice":                 "--gpsd",
	"gpsScanFrequency":          "--gpsf",
	"logDiskDirectory":          "--ld",
	"logDiskConsumptionLimit":   "--l",
	"logFileCount":              "--lc",
	"logLevel":                  "--ll",
	"memoryConsumptionLimit":    "--m",
	"networkInterface":          "--n",
	"pruningFrequency":          "--pf",
	"processorConsumptionLimit": "--p",
	"scanDevicesFreq":           "--sd",
	"secureMode":                "--sec",
	"statusFrequency":           "--sf",
	"timeZone":                  "--tz",
	"upgradeScanFrequency":      "--uf",
	"watchdogEnabled":           "--wd",
}

// buildBootstrapProfileFromAgentSpec collects only spec.config fields explicitly set in
// the deploy agent YAML. Template defaults are not included.
func buildBootstrapProfileFromAgentSpec(arch string, latitude, longitude float64, cfg *client.AgentConfiguration) map[string]interface{} {
	profile := make(map[string]interface{})

	if arch != "" && arch != "auto" {
		profile["arch"] = arch
	}
	if latitude != 0 || longitude != 0 {
		profile["gpsCoordinates"] = fmt.Sprintf("%g,%g", latitude, longitude)
	}
	if cfg == nil {
		return profile
	}

	if cfg.ContainerEngine != nil {
		engine := strings.TrimSpace(*cfg.ContainerEngine)
		if engine != "" && !strings.EqualFold(engine, "edgelet") {
			profile["containerEngine"] = engine
			profile["containerEngineUrl"] = resolveContainerEngineURL(engine, cfg)
		}
	}
	if cfg.ContainerEngineURL != nil {
		if url := strings.TrimSpace(*cfg.ContainerEngineURL); url != "" {
			profile["containerEngineUrl"] = url
		}
	}
	if cfg.NetworkInterface != nil && *cfg.NetworkInterface != "" {
		profile["networkInterface"] = *cfg.NetworkInterface
	}
	if cfg.DiskLimit != nil {
		profile["diskConsumptionLimit"] = formatInt(*cfg.DiskLimit)
	}
	if cfg.DiskDirectory != nil && *cfg.DiskDirectory != "" {
		profile["diskDirectory"] = *cfg.DiskDirectory
	}
	if cfg.MemoryLimit != nil {
		profile["memoryConsumptionLimit"] = formatInt(*cfg.MemoryLimit)
	}
	if cfg.CPULimit != nil {
		profile["processorConsumptionLimit"] = formatInt(*cfg.CPULimit)
	}
	if cfg.LogLimit != nil {
		profile["logDiskConsumptionLimit"] = formatInt(*cfg.LogLimit)
	}
	if cfg.LogDirectory != nil && *cfg.LogDirectory != "" {
		profile["logDiskDirectory"] = *cfg.LogDirectory
	}
	if cfg.LogFileCount != nil {
		profile["logFileCount"] = formatInt(*cfg.LogFileCount)
	}
	if cfg.StatusFrequency != nil {
		profile["statusFrequency"] = formatFloat(*cfg.StatusFrequency)
	}
	if cfg.ChangeFrequency != nil {
		profile["changeFrequency"] = formatFloat(*cfg.ChangeFrequency)
	}
	if cfg.DeviceScanFrequency != nil {
		profile["scanDevicesFreq"] = formatFloat(*cfg.DeviceScanFrequency)
	}
	if cfg.WatchdogEnabled != nil {
		profile["watchdogEnabled"] = boolToOnOff(*cfg.WatchdogEnabled)
	}
	if cfg.GpsMode != nil && *cfg.GpsMode != "" {
		profile["gps"] = *cfg.GpsMode
	}
	if cfg.GpsDevice != nil && *cfg.GpsDevice != "" {
		profile["gpsDevice"] = *cfg.GpsDevice
	}
	if cfg.GpsScanFrequency != nil {
		profile["gpsScanFrequency"] = formatFloat(*cfg.GpsScanFrequency)
	}
	if cfg.EdgeGuardFrequency != nil {
		profile["edgeGuardFreq"] = formatFloat(*cfg.EdgeGuardFrequency)
	}
	if cfg.PruningFrequency != nil {
		profile["pruningFrequency"] = formatFloat(*cfg.PruningFrequency)
	}
	if cfg.AvailableDiskThreshold != nil {
		profile["availableDiskThreshold"] = formatFloat(*cfg.AvailableDiskThreshold)
	}
	if cfg.LogLevel != nil && *cfg.LogLevel != "" {
		profile["logLevel"] = normalizeLogLevel(*cfg.LogLevel)
	}
	if cfg.TimeZone != "" {
		profile["timeZone"] = cfg.TimeZone
	}
	return profile
}

// BuildEdgeletBootstrapConfigCommand renders edgelet config CLI args from agent spec for desktop container bootstrap.
func BuildEdgeletBootstrapConfigCommand(arch string, latitude, longitude float64, agentCfg *client.AgentConfiguration) (string, error) {
	profile := buildBootstrapProfileFromAgentSpec(arch, latitude, longitude, agentCfg)
	if len(profile) == 0 {
		return "", nil
	}

	args := []string{"edgelet", "config"}
	for key, raw := range profile {
		if _, skip := edgeletBootstrapConfigSkipKeys[key]; skip {
			continue
		}
		flag, ok := edgeletProfileKeyToConfigFlag[key]
		if !ok {
			continue
		}
		value := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if value == "" {
			continue
		}
		args = append(args, flag, value)
	}
	if len(args) == 2 {
		return "", nil
	}
	return shellJoinArgs(args), nil
}

func (cfg EdgeletInstallConfig) bootstrapConfigCommand() (string, error) {
	arch := cfg.Arch
	latitude := 0.0
	longitude := 0.0
	var agentCfg *client.AgentConfiguration
	if cfg.Runtime != nil {
		agentCfg = &cfg.Runtime.Agent
		latitude = cfg.Runtime.Latitude
		longitude = cfg.Runtime.Longitude
		if cfg.Runtime.Arch != "" {
			arch = cfg.Runtime.Arch
		}
	}
	return BuildEdgeletBootstrapConfigCommand(arch, latitude, longitude, agentCfg)
}

func runtimeConfigDir(paths EdgeletPlatformPaths) string {
	if paths.ConfigDir != "" {
		return paths.ConfigDir
	}
	return filepath.Dir(paths.ConfigFile)
}

func ensureLocalRuntimeConfigDir(paths EdgeletPlatformPaths, useSudo bool) error {
	dir := runtimeConfigDir(paths)
	if !useSudo {
		return os.MkdirAll(dir, util.DirPerm)
	}
	if _, err := util.Exec("", "sudo", "mkdir", "-p", dir); err != nil {
		return err
	}
	_, err := util.Exec("", "sudo", "chmod", "755", dir)
	return err
}

func localRuntimeFileExists(path string, useSudo bool) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	if !useSudo {
		return false, err
	}
	// install.sh may create /etc/edgelet mode 750; unprivileged stat returns EACCES.
	if _, err := util.Exec("", "sudo", "test", "-f", path); err == nil {
		return true, nil
	}
	return false, nil
}

func writeLocalFileIfMissing(path string, content []byte, perm os.FileMode, useSudo bool) error {
	exists, err := localRuntimeFileExists(path, useSudo)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), util.DirPerm); err != nil && !useSudo {
		return err
	}

	if !useSudo {
		return os.WriteFile(path, content, perm)
	}

	tmp, err := os.CreateTemp("", "edgelet-runtime-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(content); err != nil {
		util.IgnoreClose(tmp)
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if _, err := util.Exec("", "sudo", "mkdir", "-p", filepath.Dir(path)); err != nil {
		return err
	}
	mode := fmt.Sprintf("%04o", perm)
	if _, err := util.Exec("", "sudo", "install", "-m", mode, tmpPath, path); err != nil {
		return err
	}
	return nil
}

func materializeEdgeletRuntimeTo(paths EdgeletPlatformPaths, hostOS, arch string, spec *EdgeletRuntimeSpec, useSudo bool) error {
	if err := ensureLocalRuntimeConfigDir(paths, useSudo); err != nil {
		return fmt.Errorf("prepare edgelet config dir: %w", err)
	}

	var agentCfg *client.AgentConfiguration
	latitude := 0.0
	longitude := 0.0
	if spec != nil {
		agentCfg = &spec.Agent
		latitude = spec.Latitude
		longitude = spec.Longitude
		if spec.Arch != "" {
			arch = spec.Arch
		}
	}

	configYAML, err := BuildEdgeletConfigYAML(hostOS, arch, latitude, longitude, agentCfg)
	if err != nil {
		return err
	}
	if err := writeLocalFileIfMissing(paths.ConfigFile, configYAML, 0o640, useSudo); err != nil {
		return fmt.Errorf("write edgelet config: %w", err)
	}

	sampleCA, err := loadEdgeletSampleCA()
	if err != nil {
		return err
	}
	if err := writeLocalFileIfMissing(paths.CertFile, sampleCA, 0o644, useSudo); err != nil {
		return fmt.Errorf("write edgelet sample CA: %w", err)
	}
	return nil
}

// MaterializeEdgeletRuntime writes edgelet config and sample CA on the local host when files are missing.
func MaterializeEdgeletRuntime(hostOS string, spec *EdgeletRuntimeSpec) error {
	useSudo := normalizeEdgeletHostOS(hostOS) != "windows"
	return materializeEdgeletRuntimeTo(EdgeletPaths(hostOS), hostOS, "", spec, useSudo)
}

// EdgeletProvisionCommands builds edgelet config/provision command strings for a controller endpoint.
//
//nolint:revive // command is intentionally unexported; callers are in this package
func EdgeletProvisionCommands(controllerEndpoint, key, caCert string, useSudo bool) ([]command, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("provisioning key is required")
	}

	prefix := ""
	if useSudo {
		prefix = "sudo "
	}

	controllerBaseURL, err := util.GetBaseURL(controllerEndpoint)
	if err != nil {
		return nil, err
	}

	cmds := []command{
		{
			cmd: prefix + "edgelet config --a " + controllerBaseURL.String(),
			msg: "Configuring edgelet with Controller URL " + controllerBaseURL.String(),
		},
	}
	if strings.TrimSpace(caCert) != "" {
		cmds = append(cmds, command{
			cmd: prefix + "edgelet config cert " + caCert,
			msg: "Configuring edgelet with Controller CA certificate",
		})
	}
	cmds = append(cmds, command{
		cmd: prefix + "edgelet provision " + key,
		msg: "Provisioning edgelet with Controller",
	})
	return cmds, nil
}
