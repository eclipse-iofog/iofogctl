package deployairgap

import (
	"context"
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

// AgentAirgapResult holds artifacts produced during agent airgap preparation.
type AgentAirgapResult struct {
	Platform      string
	Options       AirgapTransferOptions
	RemoteBinPath string
	ImageList     []string
}

// PrepareAgentAirgap resolves platform, options, and image refs for an agent airgap deploy.
func PrepareAgentAirgap(namespace string, agent *rsc.RemoteAgent, controlPlane *rsc.RemoteControlPlane, isInitial bool) (AgentAirgapResult, error) {
	if agent == nil || agent.Config == nil {
		return AgentAirgapResult{}, fmt.Errorf("agent configuration is required for airgap deployment")
	}
	if err := ValidateAirgapRequirements(agent.Config); err != nil {
		return AgentAirgapResult{}, err
	}

	platform, err := ResolvePlatform(agent.Config.Arch)
	if err != nil {
		return AgentAirgapResult{}, fmt.Errorf("failed to resolve platform: %w", err)
	}

	opts, err := AirgapTransferOptionsFromConfig(agent.Config)
	if err != nil {
		return AgentAirgapResult{}, fmt.Errorf("failed to resolve airgap transfer options: %w", err)
	}

	images, err := CollectAgentImages(namespace, agent, controlPlane, isInitial)
	if err != nil {
		return AgentAirgapResult{}, fmt.Errorf("failed to collect agent images: %w", err)
	}

	imageList, err := CollectAgentAirgapImages(images, platform, opts.DeploymentType)
	if err != nil {
		return AgentAirgapResult{}, fmt.Errorf("failed to build agent airgap image list: %w", err)
	}

	return AgentAirgapResult{
		Platform:  platform,
		Options:   opts,
		ImageList: imageList,
	}, nil
}

// TransferAgentAirgapBinary caches and SCPs the edgelet binary when deploymentType is native.
func TransferAgentAirgapBinary(ctx context.Context, namespace, host string, ssh *rsc.SSH, platform string) (string, error) {
	return EnsureAndTransferEdgeletBinary(ctx, namespace, host, platform, ssh)
}

// TransferAgentAirgapImages transfers and loads container images using the airgap load matrix.
func TransferAgentAirgapImages(ctx context.Context, namespace, host string, ssh *rsc.SSH, platform string, opts AirgapTransferOptions, images []string) error {
	return TransferAirgapImages(ctx, namespace, host, ssh, platform, opts, images)
}
