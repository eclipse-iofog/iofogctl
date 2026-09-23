package describe

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func MapClientMicroserviceToDeployMicroservice(msvc *client.MicroserviceInfo, clt *client.Client) (*apps.Microservice, *apps.MicroserviceStatusInfo, *apps.MicroserviceExecStatusInfo, error) {
	agent, err := clt.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return nil, nil, nil, err
	}
	var catalogItem *client.CatalogItemInfo
	if msvc.CatalogItemID != 0 {
		catalogItem, err = clt.GetCatalogItem(msvc.CatalogItemID)
		if err != nil {
			httpErr := &client.HTTPError{}
			if errors.As(err, &httpErr) {
				catalogItem = nil
			} else {
				return nil, nil, nil, err
			}
		}
	}

	applicationName := msvc.Application
	// if msvc.Application == "" {
	// 	if msvc.ApplicationName != "" {
	// 		// Legacy
	// 		application, err := clt.GetApplicationByName(msvc.ApplicationName)
	// 		if err != nil {
	// 			return nil, nil, nil, err
	// 		}
	// 		applicationName = application.Name
	// 	}
	// }

	return constructMicroservice(msvc, agent.Name, applicationName, catalogItem)
}

// func MapClientMicroserviceStatusToDeployMicroserviceStatus(msvc *client.MicroserviceInfo, clt *client.Client) (*apps.MicroserviceStatusInfo, *apps.MicroserviceExecStatusInfo, error) {
// 	msvcStatus := new(apps.MicroserviceStatusInfo)
// 	msvcStatus.Status = msvc.Status.Status
// 	msvcStatus.StartTime = msvc.Status.StartTime
// 	msvcStatus.OperatingDuration = msvc.Status.OperatingDuration
// 	msvcStatus.MemoryUsage = msvc.Status.MemoryUsage
// 	msvcStatus.CPUUsage = msvc.Status.CPUUsage
// 	msvcStatus.ContainerID = msvc.Status.ContainerID
// 	msvcStatus.Percentage = msvc.Status.Percentage
// 	msvcStatus.IPAddress = msvc.Status.IPAddress
// 	msvcStatus.ErrorMessage = msvc.Status.ErrorMessage
// 	msvcStatus.ExecSessionIDs = msvc.Status.ExecSessionIDs

// 	msvcExecStatus := new(apps.MicroserviceExecStatusInfo)
// 	msvcExecStatus.Status = msvc.ExecStatus.Status
// 	msvcExecStatus.ExecSessionID = msvc.ExecStatus.ExecSessionID

// 	return msvcStatus, msvcExecStatus, nil
// }

// FormatMicroserviceStatus formats microservice status for human-readable output.
func FormatMicroserviceStatus(status *apps.MicroserviceStatusInfo) yaml.MapSlice {
	formatted := yaml.MapSlice{}
	formatted = appendItem(formatted, "status", status.Status)
	formatted = appendItem(formatted, "containerId", status.ContainerID)
	if status.PodID != "" {
		formatted = appendItem(formatted, "podId", status.PodID)
	}
	formatted = appendItem(formatted, "percentage", status.Percentage)
	formatted = appendItem(formatted, "healthStatus", status.HealthStatus)
	formatted = appendItem(formatted, "ipAddress", status.IPAddress)
	if status.StartTime > 0 {
		formatted = appendItem(formatted, "startTime", time.Unix(status.StartTime/1000, (status.StartTime%1000)*1000000).Format(time.RFC3339))
	}
	if status.OperatingDuration > 0 {
		duration := time.Duration(status.OperatingDuration) * time.Millisecond
		formatted = appendItem(formatted, "operatingDuration", util.FormatDuration(duration))
	}
	if status.CPUUsage > 0 {
		formatted = appendItem(formatted, "cpuUsage", formatCPUCores(status.CPUUsage))
	}
	if status.MemoryUsage > 0 {
		formatted = appendItem(formatted, "memoryUsage", formatBytesAuto(status.MemoryUsage))
	}
	formatted = appendItem(formatted, "restartCount", status.RestartCount)
	formatted = appendItem(formatted, "errorMessage", status.ErrorMessage)
	formatted = appendItem(formatted, "lastError", status.LastError)
	// lastErrorAt is Unix milliseconds; keep 0 when LastError is empty (older Controllers).
	if status.LastErrorAt > 0 {
		formatted = appendItem(formatted, "lastErrorAt", time.Unix(status.LastErrorAt/1000, (status.LastErrorAt%1000)*1000000).Format(time.RFC3339))
	} else {
		formatted = appendItem(formatted, "lastErrorAt", int64(0))
	}
	formatted = appendItem(formatted, "execSessionIds", status.ExecSessionIDs)
	return formatted
}

// FormatMicroserviceExecStatus formats microservice exec status for human-readable output.
func FormatMicroserviceExecStatus(execStatus *apps.MicroserviceExecStatusInfo) yaml.MapSlice {
	return yaml.MapSlice{
		{Key: "status", Value: execStatus.Status},
		{Key: "execSessionId", Value: execStatus.ExecSessionID},
	}
}

// FormatMicroserviceDescribeStatus nests status and execStatus in stable order.
func FormatMicroserviceDescribeStatus(status *apps.MicroserviceStatusInfo, execStatus *apps.MicroserviceExecStatusInfo) yaml.MapSlice {
	return yaml.MapSlice{
		{Key: "status", Value: FormatMicroserviceStatus(status)},
		{Key: "execStatus", Value: FormatMicroserviceExecStatus(execStatus)},
	}
}

func constructMicroservice(msvcInfo *client.MicroserviceInfo, agentName, appName string, catalogItem *client.CatalogItemInfo) (msvc *apps.Microservice, status *apps.MicroserviceStatusInfo, execStatus *apps.MicroserviceExecStatusInfo, err error) {
	msvc = new(apps.Microservice)
	msvc.UUID = msvcInfo.UUID
	msvc.Name = msvcInfo.Name
	msvc.Agent = apps.MicroserviceAgent{
		Name: agentName,
	}
	var armImage, amd64Image, riscv64Image, arm64Image string
	var msvcImages []client.CatalogImage
	if catalogItem != nil {
		msvcImages = catalogItem.Images
	} else {
		msvcImages = msvcInfo.Images
	}
	for _, image := range msvcImages {
		switch client.ArchIDToName[image.ArchID] {
		case "amd64":
			amd64Image = image.ContainerImage
		case "arm64":
			armImage = image.ContainerImage
		case "riscv64":
			riscv64Image = image.ContainerImage
		case "arm":
			armImage = image.ContainerImage
		default:
		}
	}
	var registryID int
	var imgArray []client.CatalogImage
	if catalogItem != nil {
		registryID = catalogItem.RegistryID
		imgArray = catalogItem.Images
	} else {
		registryID = msvcInfo.RegistryID
		imgArray = msvcInfo.Images
	}
	images := apps.MicroserviceImages{
		CatalogID: msvcInfo.CatalogItemID,
		AMD64:     amd64Image,
		ARM64:     arm64Image,
		RISCV64:   riscv64Image,
		ARM:       armImage,
		Registry:  apps.RegistryRef(registryID),
	}
	for _, img := range imgArray {
		switch img.ArchID {
		case 1:
			images.AMD64 = img.ContainerImage
		case 2:
			images.ARM64 = img.ContainerImage
		case 3:
			images.RISCV64 = img.ContainerImage
		case 4:
			images.ARM = img.ContainerImage
		}
	}
	volumes := mapVolumes(msvcInfo.Volumes)
	envs := mapEnvs(msvcInfo.Env)
	extraHosts := mapExtraHosts(msvcInfo.ExtraHosts)
	msvc.Images = &images
	var healthCheck apps.MicroserviceHealthCheck
	var hasHealthCheck bool
	// Fix 1: Check if HealthCheck has a Test field (assuming it's a struct, not pointer)
	if msvcInfo.HealthCheck.Test != nil {
		hasHealthCheck = false
		if msvcInfo.HealthCheck.Test != nil {
			healthCheck.Test = msvcInfo.HealthCheck.Test
			hasHealthCheck = true
		}

		// Only set fields that were present in the original JSON
		if msvcInfo.HealthCheck.Interval != nil {
			healthCheck.Interval = msvcInfo.HealthCheck.Interval
		}
		if msvcInfo.HealthCheck.Timeout != nil {
			healthCheck.Timeout = msvcInfo.HealthCheck.Timeout
		}
		if msvcInfo.HealthCheck.Retries != nil {
			healthCheck.Retries = msvcInfo.HealthCheck.Retries
		}
		if msvcInfo.HealthCheck.StartPeriod != nil {
			healthCheck.StartPeriod = msvcInfo.HealthCheck.StartPeriod
		}
		if msvcInfo.HealthCheck.StartInterval != nil {
			healthCheck.StartInterval = msvcInfo.HealthCheck.StartInterval
		}
	}
	var config apps.ArbitraryJSON
	if err := config.UnmarshalJSON([]byte(msvcInfo.Config)); err != nil {
		return msvc, nil, nil, err
	}
	msvc.Config = config
	var annotations apps.ArbitraryJSON
	if err := annotations.UnmarshalJSON([]byte(msvcInfo.Annotations)); err != nil {
		return msvc, nil, nil, err
	}
	msvc.Container.Annotations = annotations
	msvc.Container.HostNetworkMode = msvcInfo.HostNetworkMode
	msvc.Container.IsPrivileged = msvcInfo.IsPrivileged
	msvc.Container.PidMode = msvcInfo.PidMode
	msvc.Container.IpcMode = msvcInfo.IpcMode
	msvc.Container.Runtime = msvcInfo.Runtime
	msvc.Container.Platform = msvcInfo.Platform
	msvc.Container.RunAsUser = msvcInfo.RunAsUser
	msvc.Container.RunAsGroup = msvcInfo.RunAsGroup
	msvc.Container.ReadOnlyRootFilesystem = msvcInfo.ReadOnlyRootFilesystem
	msvc.Container.CdiDevices = msvcInfo.CdiDevices
	msvc.Container.CapAdd = msvcInfo.CapAdd
	msvc.Container.CapDrop = msvcInfo.CapDrop
	// Commands is the merged cmd/commands list from GET (commands wins when both are present).
	msvc.Container.Commands = msvcInfo.Commands
	if len(msvcInfo.CommandsAlias) > 0 {
		msvc.Container.Commands = msvcInfo.CommandsAlias
	}
	msvc.Container.Entrypoint = msvcInfo.Entrypoint
	msvc.Container.WorkingDir = msvcInfo.WorkingDir
	msvc.Container.Ports = mapPorts(msvcInfo.Ports)
	msvc.Container.Volumes = &volumes
	msvc.Container.Env = &envs
	msvc.Container.ExtraHosts = &extraHosts
	msvc.Container.CPUSetCpus = msvcInfo.CPUSetCpus
	msvc.Container.MemoryLimit = &msvcInfo.MemoryLimit
	if msvcInfo.Cpus != 0 {
		cpus := msvcInfo.Cpus
		msvc.Container.CPUs = &cpus
	}
	if msvcInfo.MemoryReservation != 0 {
		reservation := msvcInfo.MemoryReservation
		msvc.Container.MemoryReservation = &reservation
	}
	if msvcInfo.MemorySwap != 0 {
		swap := msvcInfo.MemorySwap
		msvc.Container.MemorySwap = &swap
	}
	if msvcInfo.ShmSize != 0 {
		shm := msvcInfo.ShmSize
		msvc.Container.ShmSize = &shm
	}
	msvc.Container.Sysctls = msvcInfo.Sysctls
	msvc.Container.Ulimits = mapUlimits(msvcInfo.Ulimits)
	msvc.Container.Devices = mapDevices(msvcInfo.Devices)
	msvc.Container.Tmpfs = mapTmpfs(msvcInfo.Tmpfs)
	if hasHealthCheck {
		msvc.Container.HealthCheck = &healthCheck
	}
	msvc.Models = mapMicroserviceCatalog(msvcInfo.Models)
	msvc.Knowledge = mapMicroserviceKnowledgeCatalog(msvcInfo.Knowledge)
	if msvcInfo.NatsConfig != nil {
		msvc.NatsConfig = &apps.MicroserviceNatsConfig{
			NatsAccess: msvcInfo.NatsConfig.NatsAccess,
			NatsRule:   msvcInfo.NatsConfig.NatsRule,
		}
	}
	msvc.Schedule = msvcInfo.Schedule
	msvc.Application = appName
	if msvcInfo.ServiceAccount != nil {
		msvc.ServiceAccount = &apps.MicroserviceServiceAccountRef{
			RoleRef: apps.RoleRef{
				Kind:     msvcInfo.ServiceAccount.RoleRef.Kind,
				Name:     msvcInfo.ServiceAccount.RoleRef.Name,
				APIGroup: msvcInfo.ServiceAccount.RoleRef.APIGroup,
			},
		}
	}
	status = new(apps.MicroserviceStatusInfo)

	status.Status = msvcInfo.Status.Status
	status.StartTime = msvcInfo.Status.StartTime
	status.OperatingDuration = msvcInfo.Status.OperatingDuration
	status.MemoryUsage = msvcInfo.Status.MemoryUsage
	status.CPUUsage = msvcInfo.Status.CPUUsage
	status.ContainerID = msvcInfo.Status.ContainerID
	status.Percentage = msvcInfo.Status.Percentage
	status.ErrorMessage = msvcInfo.Status.ErrorMessage
	status.LastError = msvcInfo.Status.LastError
	status.LastErrorAt = msvcInfo.Status.LastErrorAt
	status.RestartCount = msvcInfo.Status.RestartCount
	status.IPAddress = msvcInfo.Status.IPAddress
	status.ExecSessionIDs = msvcInfo.Status.ExecSessionIDs
	status.HealthStatus = msvcInfo.Status.HealthStatus
	status.PodID = msvcInfo.Status.PodID
	execStatus = new(apps.MicroserviceExecStatusInfo)
	execStatus.Status = msvcInfo.ExecStatus.Status
	execStatus.ExecSessionID = msvcInfo.ExecStatus.ExecSessionID

	return msvc, status, execStatus, err
}

func mapPort(in *client.MicroservicePortMappingInfo) (out *apps.MicroservicePortMapping) {
	if in == nil {
		return nil
	}
	return &apps.MicroservicePortMapping{
		Internal: in.Internal,
		External: in.External,
		Protocol: in.Protocol,
	}
}

func mapPorts(in []client.MicroservicePortMappingInfo) (out []apps.MicroservicePortMapping) {
	for idx := range in {
		port := mapPort(&in[idx])
		if port != nil {
			out = append(out, *port)
		}
	}
	return
}

func mapVolumes(in []client.MicroserviceVolumeMappingInfo) (out []apps.MicroserviceVolumeMapping) {
	for _, vol := range in {
		out = append(out, apps.MicroserviceVolumeMapping(vol))
	}
	return
}

func mapEnvs(in []client.MicroserviceEnvironmentInfo) (out []apps.MicroserviceEnvironment) {
	for _, env := range in {
		out = append(out, apps.MicroserviceEnvironment(env))
	}
	return
}

func mapExtraHosts(in []client.MicroserviceExtraHost) (out []apps.MicroserviceExtraHost) {
	for _, eH := range in {
		out = append(out, apps.MicroserviceExtraHost(eH))
	}
	return
}

func mapMicroserviceCatalog(in *client.MicroserviceCatalog) *apps.MicroserviceCatalog {
	if in == nil {
		return nil
	}
	out := &apps.MicroserviceCatalog{
		BindPath:    in.BindPath,
		Permissions: in.Permissions,
	}
	if len(in.Items) > 0 {
		out.Items = make([]apps.MicroserviceCatalogItem, len(in.Items))
		for i, item := range in.Items {
			out.Items[i] = apps.MicroserviceCatalogItem{Name: item.Name}
		}
	}
	return out
}

func mapMicroserviceKnowledgeCatalog(in *client.KnowledgeCatalog) *apps.KnowledgeCatalog {
	if in == nil {
		return nil
	}
	out := &apps.KnowledgeCatalog{
		BindPath:    in.BindPath,
		Permissions: in.Permissions,
	}
	if len(in.Items) > 0 {
		out.Items = make([]apps.KnowledgeCatalogItem, len(in.Items))
		for i, item := range in.Items {
			out.Items[i] = apps.KnowledgeCatalogItem{Name: item.Name}
		}
	}
	return out
}

func mapUlimits(in map[string]client.ContainerUlimit) map[string]apps.MicroserviceUlimit {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]apps.MicroserviceUlimit, len(in))
	for k, u := range in {
		out[k] = apps.MicroserviceUlimit{Soft: u.Soft, Hard: u.Hard}
	}
	return out
}

func mapDevices(in []client.ContainerDevice) []apps.MicroserviceDevice {
	if len(in) == 0 {
		return nil
	}
	out := make([]apps.MicroserviceDevice, len(in))
	for i, d := range in {
		out[i] = apps.MicroserviceDevice{
			HostPath:      d.HostPath,
			ContainerPath: d.ContainerPath,
			Permissions:   d.Permissions,
		}
	}
	return out
}

func mapTmpfs(in []client.ContainerTmpfs) []apps.MicroserviceTmpfs {
	if len(in) == 0 {
		return nil
	}
	out := make([]apps.MicroserviceTmpfs, len(in))
	for i, t := range in {
		out[i] = apps.MicroserviceTmpfs{
			ContainerPath: t.ContainerPath,
			Mode:          t.Mode,
		}
		if t.Size != 0 {
			size := t.Size
			out[i].Size = &size
		}
	}
	return out
}

// FormatAgentStatus formats agent status for human-readable output in a stable field order.
func FormatAgentStatus(status rsc.AgentStatus) yaml.MapSlice {
	formatted := yaml.MapSlice{}

	formatted = appendItem(formatted, "version", status.Version)
	formatted = appendItem(formatted, "daemonStatus", status.DaemonStatus)
	formatted = appendItem(formatted, "securityStatus", status.SecurityStatus)
	formatted = appendItem(formatted, "securityViolationInfo", status.SecurityViolationInfo)
	formatted = appendItem(formatted, "warningMessage", status.WarningMessage)
	formatted = appendItem(formatted, "gpsStatus", status.GpsStatus)
	formatted = appendItem(formatted, "ipAddress", status.IPAddress)
	formatted = appendItem(formatted, "ipAddressExternal", status.IPAddressExternal)
	if status.LastActive > 0 {
		formatted = appendItem(formatted, "lastActive", time.Unix(status.LastActive/1000, (status.LastActive%1000)*1000000).UTC().Format(time.RFC3339))
	}
	if status.LastStatusTimeMsUTC > 0 {
		formatted = appendItem(formatted, "lastStatusTime", time.Unix(status.LastStatusTimeMsUTC/1000, (status.LastStatusTimeMsUTC%1000)*1000000).UTC().Format(time.RFC3339))
	}
	if status.LastCommandTimeMsUTC > 0 {
		formatted = appendItem(formatted, "lastCommandTime", time.Unix(status.LastCommandTimeMsUTC/1000, (status.LastCommandTimeMsUTC%1000)*1000000).Format(time.RFC3339))
	} else {
		formatted = appendItem(formatted, "lastCommandTime", "Never")
	}
	if status.UptimeMs > 0 {
		uptime := time.Duration(status.UptimeMs) * time.Millisecond
		formatted = appendItem(formatted, "uptime", util.FormatDuration(uptime))
	}

	// memoryUsage is MiB (binary); diskUsage is GiB (decimal); cpuUsage is 100 = 1 core.
	if status.CPUUsage > 0 {
		formatted = appendItem(formatted, "cpuUsage", formatCPUCores(status.CPUUsage))
	}
	if status.MemoryUsage > 0 {
		formatted = appendItem(formatted, "memoryUsage", formatBytesAuto(status.MemoryUsage*1024*1024))
	}
	if status.DiskUsage > 0 {
		formatted = appendItem(formatted, "diskUsage", formatBytesAuto(status.DiskUsage*1_000_000_000))
	}
	formatted = appendItem(formatted, "cpuViolation", status.CPUViolation)
	formatted = appendItem(formatted, "memoryViolation", status.MemoryViolation)
	formatted = appendItem(formatted, "diskViolation", status.DiskViolation)

	if hostMetricsReported(status) {
		if status.SystemOs != "" {
			formatted = appendItem(formatted, "systemOs", status.SystemOs)
		}
		if status.SystemOsVersion != "" {
			formatted = appendItem(formatted, "systemOsVersion", status.SystemOsVersion)
		}
		if status.SystemKernelVersion != "" {
			formatted = appendItem(formatted, "systemKernelVersion", status.SystemKernelVersion)
		}
		if status.SystemCpus != 0 {
			formatted = appendItem(formatted, "systemCpus", status.SystemCpus)
		}
		if status.SystemTotalCPU > 0 {
			formatted = appendItem(formatted, "systemTotalCPU", fmt.Sprintf("%.2f %%", status.SystemTotalCPU))
		}
		if status.SystemTotalMemory > 0 {
			formatted = appendItem(formatted, "systemTotalMemory", formatBytesAuto(float64(status.SystemTotalMemory)))
		}
		if status.SystemAvailableMemory > 0 {
			formatted = appendItem(formatted, "systemAvailableMemory", formatBytesAuto(float64(status.SystemAvailableMemory)))
		}
		if status.SystemTotalDisk > 0 {
			formatted = appendItem(formatted, "systemTotalDisk", formatBytesAuto(float64(status.SystemTotalDisk)))
		}
		if status.SystemAvailableDisk > 0 {
			formatted = appendItem(formatted, "systemAvailableDisk", formatBytesAuto(float64(status.SystemAvailableDisk)))
		}
	}

	formatted = appendItem(formatted, "availableRuntimes", status.AvailableRuntimes)
	if parsed, ok := parseJSONBlob(status.RuntimeClasses); ok {
		formatted = appendItem(formatted, "runtimeClasses", parsed)
	}
	formatted = appendItem(formatted, "runtimeAgentPhase", status.RuntimeAgentPhase)
	if parsed, ok := parseJSONBlob(status.AvailableCdiDevices); ok {
		formatted = appendItem(formatted, "availableCdiDevices", parsed)
	}
	formatted = appendItem(formatted, "controlPlaneQuiesced", status.ControlPlaneQuiesced)
	if status.PlatformStatus != nil {
		formatted = appendItem(formatted, "platformStatus", formatPlatformStatus(status.PlatformStatus))
	}

	formatted = appendItem(formatted, "activeModels", status.ActiveModels)
	if parsed, ok := parseJSONBlob(status.ModelStatus); ok {
		formatted = appendItem(formatted, "modelStatus", parsed)
	}
	if status.ModelLastUpdate > 0 {
		formatted = appendItem(formatted, "modelLastUpdate", formatUnixMillisUTC(status.ModelLastUpdate))
	}
	formatted = appendItem(formatted, "activeKnowledge", status.ActiveKnowledge)
	if parsed, ok := parseJSONBlob(status.KnowledgeStatus); ok {
		formatted = appendItem(formatted, "knowledgeStatus", parsed)
	}
	if status.KnowledgeLastUpdate > 0 {
		formatted = appendItem(formatted, "knowledgeLastUpdate", formatUnixMillisUTC(status.KnowledgeLastUpdate))
	}

	formatted = appendItem(formatted, "repositoryStatus", status.RepositoryStatus)
	formatted = appendItem(formatted, "isReadyToUpgrade", status.IsReadyToUpgrade)
	formatted = appendItem(formatted, "isReadyToRollback", status.IsReadyToRollback)
	formatted = appendItem(formatted, "tunnel", status.Tunnel)
	formatted = appendItem(formatted, "volumeMounts", status.VolumeMounts)

	return formatted
}

func hostMetricsReported(status rsc.AgentStatus) bool {
	return status.SystemCpus != 0 ||
		status.SystemTotalMemory != 0 ||
		status.SystemAvailableMemory != 0 ||
		status.SystemTotalDisk != 0 ||
		status.SystemAvailableDisk != 0 ||
		status.SystemTotalCPU != 0 ||
		status.SystemOs != "" ||
		status.SystemOsVersion != "" ||
		status.SystemKernelVersion != ""
}

func appendItem(dst yaml.MapSlice, key string, value interface{}) yaml.MapSlice {
	return append(dst, yaml.MapItem{Key: key, Value: value})
}

func formatCPUCores(usage float64) string {
	return fmt.Sprintf("%.2f cores", usage/100)
}

func formatUnixMillisUTC(ms int64) string {
	return time.Unix(ms/1000, (ms%1000)*1_000_000).UTC().Format(time.RFC3339)
}

func parseJSONBlob(raw string) (interface{}, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw, true
	}
	if v == nil {
		return nil, false
	}
	return v, true
}

func formatPlatformStatus(ps *client.PlatformStatus) map[string]interface{} {
	if ps == nil {
		return nil
	}
	out := map[string]interface{}{
		"phase":              string(ps.Phase),
		"generation":         ps.Generation,
		"observedGeneration": ps.ObservedGeneration,
	}
	if ps.LastError != nil {
		out["lastError"] = *ps.LastError
	} else {
		out["lastError"] = nil
	}
	if ps.LastTransitionAt != nil {
		out["lastTransitionAt"] = ps.LastTransitionAt.Format(time.RFC3339Nano)
	}
	if len(ps.Conditions) > 0 {
		conditions := make([]map[string]interface{}, len(ps.Conditions))
		for i, c := range ps.Conditions {
			cond := map[string]interface{}{
				"type":   c.Type,
				"status": c.Status,
			}
			if c.Reason != "" {
				cond["reason"] = c.Reason
			}
			if c.Message != "" {
				cond["message"] = c.Message
			}
			conditions[i] = cond
		}
		out["conditions"] = conditions
	}
	return out
}

// formatBytesAuto formats bytes with automatic unit scaling (B, KB, MB, GB, etc.)
func formatBytesAuto(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.0f B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", bytes/float64(div), "KMGTPE"[exp])
}
