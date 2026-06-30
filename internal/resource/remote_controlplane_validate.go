package resource

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

const remoteControlPlaneLabel = "Remote Control Plane"

// ValidateRemoteControlPlaneMetadata rejects retired deploy YAML metadata fields.
func ValidateRemoteControlPlaneMetadata(fullYAML []byte) error {
	var doc struct {
		Metadata map[string]interface{} `yaml:"metadata"`
	}
	if err := yaml.Unmarshal(fullYAML, &doc); err != nil {
		return util.NewUnmarshalError(err.Error())
	}
	if _, ok := doc.Metadata["controlPlaneType"]; ok {
		return util.NewInputError("metadata.controlPlaneType is retired; use kind: ControlPlane")
	}
	return nil
}

// ValidateRemoteControlPlane validates a parsed RemoteControlPlane spec.
func ValidateRemoteControlPlane(cp *RemoteControlPlane) error {
	if err := validateIofogUser(remoteControlPlaneLabel, cp.IofogUser); err != nil {
		return err
	}
	if err := validateAuth(remoteControlPlaneLabel, cp.Auth); err != nil {
		return err
	}
	if err := validateEndpointMatch(remoteControlPlaneLabel, cp.Endpoint, cp.Controller.PublicUrl); err != nil {
		return err
	}
	if err := validateControllerPackage(remoteControlPlaneLabel, cp.Controller.Package); err != nil {
		return err
	}
	if err := validateRemoteDatabase(cp.Database, len(cp.Controllers)); err != nil {
		return err
	}
	if err := validateRemoteControllers(cp.Controllers); err != nil {
		return err
	}
	if cp.Airgap {
		if err := validateRemoteControlPlaneAirgapArch(cp.Controllers); err != nil {
			return err
		}
	}
	if err := validateRemoteSystemMicroservices(cp.SystemMicroservices); err != nil {
		return err
	}
	if err := validateCAField(remoteControlPlaneLabel, "ca", cp.CA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(remoteControlPlaneLabel, "routerSiteCA", cp.RouterSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(remoteControlPlaneLabel, "routerLocalCA", cp.RouterLocalCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(remoteControlPlaneLabel, "natsSiteCA", cp.NatsSiteCA); err != nil {
		return err
	}
	if err := validateSiteCertificateBlock(remoteControlPlaneLabel, "natsLocalCA", cp.NatsLocalCA); err != nil {
		return err
	}
	if err := validateControlPlaneTLS(remoteControlPlaneLabel, cp.TLS); err != nil {
		return err
	}
	if err := validateVault(remoteControlPlaneLabel, cp.Vault); err != nil {
		return err
	}
	return nil
}

func validateRemoteDatabase(db Database, controllerCount int) error {
	if controllerCount > 1 {
		if db.Provider == "" {
			return util.NewInputError("Remote Control Plane database is required when multiple controllers are configured")
		}
	}
	if db.Provider == "" {
		return nil
	}
	return validateDatabase(remoteControlPlaneLabel, db)
}

func validateRemoteControllers(controllers []RemoteController) error {
	if len(controllers) == 0 {
		return util.NewInputError("Remote Control Plane requires at least one controller")
	}
	seen := make(map[string]struct{}, len(controllers))
	for _, ctrl := range controllers {
		if err := util.IsLowerAlphanumeric("Controller", ctrl.Name); err != nil {
			return err
		}
		if _, exists := seen[ctrl.Name]; exists {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller name %q must be unique", ctrl.Name))
		}
		seen[ctrl.Name] = struct{}{}
		if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.KeyFile == "" {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q requires host, ssh.user, and ssh.keyFile", ctrl.Name))
		}
		if ctrl.SystemAgent == nil {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q requires systemAgent", ctrl.Name))
		}
		if err := validateRemoteControllerSystemAgent(ctrl.Name, ctrl.SystemAgent); err != nil {
			return err
		}
		if err := validateControlPlaneTLS(remoteControlPlaneLabel, ctrl.TLS); err != nil {
			return err
		}
	}
	return nil
}

func validateRemoteControllerSystemAgent(controllerName string, systemAgent *SystemAgentConfig) error {
	label := fmt.Sprintf("Remote Control Plane controller %q", controllerName)
	if err := validateSystemAgentPackage(label, systemAgent); err != nil {
		return err
	}
	if systemAgent.AgentConfiguration == nil {
		return nil
	}
	cfg := systemAgent.AgentConfiguration
	if cfg.Arch == nil || *cfg.Arch == "" {
		return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q systemAgent.config.arch is required when systemAgent.config is set", controllerName))
	}
	if _, ok := ArchStringToID(*cfg.Arch); !ok {
		return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q systemAgent.config.arch %q is invalid", controllerName, *cfg.Arch))
	}
	return validateSystemAgentRouterNats(label, cfg)
}

func validateRemoteControlPlaneAirgapArch(controllers []RemoteController) error {
	for _, ctrl := range controllers {
		if ctrl.SystemAgent == nil || ctrl.SystemAgent.AgentConfiguration == nil {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q requires systemAgent.config.arch when airgap is enabled", ctrl.Name))
		}
		arch := ctrl.SystemAgent.AgentConfiguration.Arch
		if arch == nil || strings.TrimSpace(*arch) == "" {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q requires systemAgent.config.arch when airgap is enabled", ctrl.Name))
		}
		if _, ok := ArchStringToID(*arch); !ok {
			return util.NewInputError(fmt.Sprintf("Remote Control Plane controller %q systemAgent.config.arch %q is invalid", ctrl.Name, *arch))
		}
	}
	return nil
}

func validateRemoteSystemMicroservices(_ install.RemoteSystemMicroservices) error {
	return nil
}
