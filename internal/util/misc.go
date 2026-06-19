package util

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func IsSystemAgent(agentConfig *rsc.AgentConfiguration) bool {
	return agentConfig != nil && agentConfig.IsSystem != nil && *agentConfig.IsSystem
}

func MakeIntPtr(value int) *int {
	return &value
}

func MakeStrPtr(value string) *string {
	return &value
}

func MakeBoolPtr(value bool) *bool {
	return &value
}
