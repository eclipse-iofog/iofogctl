package describe

import (
	"errors"
	"fmt"
	"time"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	// "github.com/eclipse-iofog/iofogctl/pkg/iofog"
	// "github.com/eclipse-iofog/iofogctl/pkg/util"
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
	if msvc.Application == "" {
		if msvc.FlowID > 0 {
			// Legacy
			flow, err := clt.GetFlowByID(msvc.FlowID)
			if err != nil {
				return nil, nil, nil, err
			}
			applicationName = flow.Name
		}
	}

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

// FormatMicroserviceStatus formats microservice status for human-readable output
func FormatMicroserviceStatus(status *apps.MicroserviceStatusInfo) map[string]interface{} {
	formatted := make(map[string]interface{})

	// Core status fields
	formatted["status"] = status.Status
	formatted["containerId"] = status.ContainerID
	formatted["percentage"] = status.Percentage
	formatted["errorMessage"] = status.ErrorMessage
	formatted["ipAddress"] = status.IPAddress
	formatted["execSessionIds"] = status.ExecSessionIDs
	formatted["healthStatus"] = status.HealthStatus

	// Format startTime as RFC3339 timestamp
	if status.StartTime > 0 {
		formatted["startTime"] = time.Unix(status.StartTime/1000, (status.StartTime%1000)*1000000).Format(time.RFC3339)
	}

	// Format operatingDuration as human-readable duration
	if status.OperatingDuration > 0 {
		duration := time.Duration(status.OperatingDuration) * time.Millisecond
		formatted["operatingDuration"] = util.FormatDuration(duration)
	}

	// Format memory usage
	if status.MemoryUsage > 0 {
		formatted["memoryUsage"] = formatBytesAuto(status.MemoryUsage)
	}

	// Format CPU usage
	if status.CPUUsage > 0 {
		formatted["cpuUsage"] = fmt.Sprintf("%.2f %%", status.CPUUsage)
	}

	return formatted
}

// FormatMicroserviceExecStatus formats microservice exec status for human-readable output
func FormatMicroserviceExecStatus(execStatus *apps.MicroserviceExecStatusInfo) map[string]interface{} {
	formatted := make(map[string]interface{})

	formatted["status"] = execStatus.Status
	formatted["execSessionId"] = execStatus.ExecSessionID

	return formatted
}

func constructMicroservice(msvcInfo *client.MicroserviceInfo, agentName, appName string, catalogItem *client.CatalogItemInfo) (msvc *apps.Microservice, status *apps.MicroserviceStatusInfo, execStatus *apps.MicroserviceExecStatusInfo, err error) {
	msvc = new(apps.Microservice)
	msvc.UUID = msvcInfo.UUID
	msvc.Name = msvcInfo.Name
	msvc.Agent = apps.MicroserviceAgent{
		Name: agentName,
	}
	var armImage, x86Image string
	var msvcImages []client.CatalogImage
	if catalogItem != nil {
		msvcImages = catalogItem.Images
	} else {
		msvcImages = msvcInfo.Images
	}
	for _, image := range msvcImages {
		switch client.AgentTypeIDAgentTypeDict[image.AgentTypeID] {
		case "x86":
			x86Image = image.ContainerImage
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
		X86:       x86Image,
		ARM:       armImage,
		Registry:  client.RegistryTypeIDRegistryTypeDict[registryID],
	}
	for _, img := range imgArray {
		switch img.AgentTypeID {
		case 1:
			images.X86 = img.ContainerImage
		case 2:
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
	msvc.Container.CdiDevices = msvcInfo.CdiDevices
	msvc.Container.CapAdd = msvcInfo.CapAdd
	msvc.Container.CapDrop = msvcInfo.CapDrop
	msvc.Container.Commands = msvcInfo.Commands
	msvc.Container.Ports = mapPorts(msvcInfo.Ports)
	msvc.Container.Volumes = &volumes
	msvc.Container.Env = &envs
	msvc.Container.ExtraHosts = &extraHosts
	msvc.Container.CpuSetCpus = msvcInfo.CpuSetCpus
	msvc.Container.MemoryLimit = &msvcInfo.MemoryLimit
	if hasHealthCheck {
		msvc.Container.HealthCheck = &healthCheck
	}
	if msvcInfo.NatsConfig != nil {
		msvc.NatsConfig = &apps.MicroserviceNatsConfig{
			NatsAccess: msvcInfo.NatsConfig.NatsAccess,
			NatsRule:   msvcInfo.NatsConfig.NatsRule,
		}
	}
	msvc.Schedule = msvcInfo.Schedule
	msvc.Application = &appName
	status = new(apps.MicroserviceStatusInfo)

	status.Status = msvcInfo.Status.Status
	status.StartTime = msvcInfo.Status.StartTime
	status.OperatingDuration = msvcInfo.Status.OperatingDuration
	status.MemoryUsage = msvcInfo.Status.MemoryUsage
	status.CPUUsage = msvcInfo.Status.CPUUsage
	status.ContainerID = msvcInfo.Status.ContainerID
	status.Percentage = msvcInfo.Status.Percentage
	status.ErrorMessage = msvcInfo.Status.ErrorMessage
	status.IPAddress = msvcInfo.Status.IPAddress
	status.ExecSessionIDs = msvcInfo.Status.ExecSessionIDs
	status.HealthStatus = msvcInfo.Status.HealthStatus
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

// FormatAgentStatus formats agent status for human-readable output
func FormatAgentStatus(status rsc.AgentStatus) map[string]interface{} {
	// Use ordered map to ensure consistent output order
	formatted := make(map[string]interface{})

	// Core status fields (ordered for consistent output)
	formatted["daemonStatus"] = status.DaemonStatus
	formatted["securityStatus"] = status.SecurityStatus
	formatted["warningMessage"] = status.WarningMessage
	formatted["securityViolationInfo"] = status.SecurityViolationInfo

	// Format timestamps
	if status.LastActive > 0 {
		formatted["lastActive"] = time.Unix(status.LastActive/1000, (status.LastActive%1000)*1000000).Format(time.RFC3339)
	}

	// Format uptime as duration
	if status.UptimeMs > 0 {
		uptime := time.Duration(status.UptimeMs) * time.Millisecond
		formatted["uptime"] = util.FormatDuration(uptime)
	}

	// Format usage - Memory/Disk are already in MiB, CPU is percentage
	if status.MemoryUsage > 0 {
		// Convert from MiB to bytes for auto-scaling
		memoryBytes := status.MemoryUsage * 1024 * 1024
		formatted["memoryUsage"] = formatBytesAuto(memoryBytes)
	}
	if status.DiskUsage > 0 {
		// Convert from MiB to bytes for auto-scaling
		diskBytes := status.DiskUsage * 1024 * 1024
		formatted["diskUsage"] = formatBytesAuto(diskBytes)
	}
	if status.CPUUsage > 0 {
		formatted["cpuUsage"] = fmt.Sprintf("%.2f %%", status.CPUUsage)
	}

	// Format system available resources
	if status.SystemAvailableMemory > 0 {
		// Convert from KB to bytes for auto-scaling
		memoryBytes := status.SystemAvailableMemory
		formatted["systemAvailableMemory"] = formatBytesAuto(memoryBytes)
	}
	if status.SystemAvailableDisk > 0 {
		// Convert from bytes to bytes for auto-scaling (already in bytes)
		formatted["systemAvailableDisk"] = formatBytesAuto(float64(status.SystemAvailableDisk))
	}

	// Add system total CPU (placeholder since not available from SDK)
	if status.SystemTotalCPU > 0 {
		formatted["systemTotalCPU"] = fmt.Sprintf("%.2f %%", status.SystemTotalCPU)
	}

	formatted["memoryViolation"] = status.MemoryViolation
	formatted["diskViolation"] = status.DiskViolation
	formatted["cpuViolation"] = status.CPUViolation
	formatted["repositoryStatus"] = status.RepositoryStatus

	// Format last status time
	if status.LastStatusTimeMsUTC > 0 {
		formatted["lastStatusTime"] = time.Unix(status.LastStatusTimeMsUTC/1000, (status.LastStatusTimeMsUTC%1000)*1000000).Format(time.RFC3339)
	}

	formatted["ipAddress"] = status.IPAddress
	formatted["ipAddressExternal"] = status.IPAddressExternal

	// Format processed messages
	if status.ProcessedMessaged > 0 {
		formatted["processedMessages"] = formatNumber(status.ProcessedMessaged)
	}

	// Format message speed
	if status.MessageSpeed > 0 {
		formatted["messageSpeed"] = fmt.Sprintf("%.1f msg/s", status.MessageSpeed)
	}

	// Format last command time
	if status.LastCommandTimeMsUTC > 0 {
		formatted["lastCommandTime"] = time.Unix(status.LastCommandTimeMsUTC/1000, (status.LastCommandTimeMsUTC%1000)*1000000).Format(time.RFC3339)
	} else {
		formatted["lastCommandTime"] = "Never"
	}

	formatted["version"] = status.Version
	formatted["isReadyToUpgrade"] = status.IsReadyToUpgrade
	formatted["isReadyToRollback"] = status.IsReadyToRollback
	formatted["tunnel"] = status.Tunnel
	formatted["volumeMounts"] = status.VolumeMounts
	formatted["gpsStatus"] = status.GpsStatus

	return formatted
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

func formatNumber(num int64) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	}
	if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	if num < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(num)/1000000000)
}
