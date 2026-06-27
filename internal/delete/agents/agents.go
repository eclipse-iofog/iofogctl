package deleteagents

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type DeleteTarget struct {
	Name  string
	Force bool
}

func CollectDeleteTargets(ns *rsc.Namespace, namespace string, force bool, excludeNames []string) ([]DeleteTarget, error) {
	if err := clientutil.SyncAgentInfo(namespace); err != nil && !rsc.IsNoControlPlaneError(err) {
		return nil, err
	}

	excluded := make(map[string]bool, len(excludeNames))
	for _, name := range excludeNames {
		excluded[name] = true
	}

	seen := make(map[string]bool)
	systemAgents := make(map[string]bool)
	targets := make([]DeleteTarget, 0)

	for _, agent := range ns.GetAgents() {
		if seen[agent.GetName()] || excluded[agent.GetName()] {
			continue
		}
		seen[agent.GetName()] = true
		if cfg := agent.GetConfig(); cfg != nil && cfg.IsSystem != nil && *cfg.IsSystem {
			systemAgents[agent.GetName()] = true
		}
		targets = append(targets, DeleteTarget{Name: agent.GetName()})
	}

	backendAgents, err := clientutil.GetBackendAgents(namespace)
	if err != nil {
		return targets, nil
	}
	for idx := range backendAgents {
		agent := &backendAgents[idx]
		if excluded[agent.Name] {
			continue
		}
		if agent.IsSystem {
			systemAgents[agent.Name] = true
		}
		if seen[agent.Name] {
			continue
		}
		seen[agent.Name] = true
		targets = append(targets, DeleteTarget{Name: agent.Name})
	}

	for i := range targets {
		targets[i].Force = force || systemAgents[targets[i].Name]
	}
	return targets, nil
}
