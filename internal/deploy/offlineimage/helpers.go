package deployofflineimage

import (
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	platformAMD64 = "linux/amd64"
	platformARM64 = "linux/arm64"
)

type agentPlan struct {
	agent    *rsc.RemoteAgent
	platform string
	engine   containerEngine
	imageRef string
}

type containerEngine string

const (
	engineDocker containerEngine = "docker"
	enginePodman containerEngine = "podman"
)

func (e containerEngine) command() string {
	return string(e)
}

func resolvePlatform(fogType *string) (string, error) {
	if fogType == nil {
		return "", util.NewInputError("Agent fog type is not configured in Controller")
	}
	value := strings.ToLower(strings.TrimSpace(*fogType))
	switch value {
	case "1", "x86", "amd64", platformAMD64:
		return platformAMD64, nil
	case "2", "arm", "arm64", platformARM64:
		return platformARM64, nil
	default:
		return "", util.NewInputError("Unsupported fog type " + *fogType)
	}
}

func resolveContainerEngine(engine *string) (containerEngine, error) {
	if engine == nil {
		return "", util.NewInputError("Agent container engine configuration is missing")
	}
	value := strings.ToLower(strings.TrimSpace(*engine))
	switch value {
	case "docker":
		return engineDocker, nil
	case "podman":
		return enginePodman, nil
	default:
		return "", util.NewInputError("Unsupported container engine " + *engine)
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if util.IsNotFoundError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "notfounderror")
}

func cleanupArtifacts(artifacts map[string]*imageArtifact) {
	for _, artifact := range artifacts {
		if artifact == nil || artifact.cleanup == nil {
			continue
		}
		if err := artifact.cleanup(); err != nil {
			util.PrintNotify(err.Error())
		}
	}
}

func sanitizeSegment(value string) string {
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
