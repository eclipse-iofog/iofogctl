package resource

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type RemoteSystemMicroservices = install.RemoteSystemMicroservices

type RemoteControlPlane struct {
	Endpoint            string                    `yaml:"endpoint,omitempty"`
	CA                  string                    `yaml:"ca,omitempty"`
	IofogUser           IofogUser                 `yaml:"iofogUser"`
	Controller          LocalControllerSpec       `yaml:"controller,omitempty"`
	Controllers         []RemoteController        `yaml:"controllers"`
	Database            Database                  `yaml:"database"`
	Auth                Auth                      `yaml:"auth"`
	RouterSiteCA        *SiteCertificate          `yaml:"routerSiteCA,omitempty"`
	RouterLocalCA       *SiteCertificate          `yaml:"routerLocalCA,omitempty"`
	NatsSiteCA          *SiteCertificate          `yaml:"natsSiteCA,omitempty"`
	NatsLocalCA         *SiteCertificate          `yaml:"natsLocalCA,omitempty"`
	SystemMicroservices RemoteSystemMicroservices `yaml:"systemMicroservices,omitempty"`
	Nats                *NatsEnabledConfig        `yaml:"nats,omitempty"`
	Events              Events                    `yaml:"events,omitempty"`
	Vault               *VaultSpec                `yaml:"vault,omitempty"`
	TLS                 *ControlPlaneTLS          `yaml:"tls,omitempty"`
	Airgap              bool                      `yaml:"airgap,omitempty"`
}

func (cp *RemoteControlPlane) GetTrustCA() string {
	return cp.CA
}

func (cp *RemoteControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp *RemoteControlPlane) UpdateUserTokens(accessToken, refreshToken string) IofogUser {
	cp.IofogUser.AccessToken = accessToken
	cp.IofogUser.RefreshToken = refreshToken

	return cp.IofogUser
}

func (cp *RemoteControlPlane) GetControllers() (controllers []Controller) {
	for idx := range cp.Controllers {
		controllers = append(controllers, cp.Controllers[idx].Clone())
	}
	return
}

func (cp *RemoteControlPlane) GetController(name string) (ret Controller, err error) {
	for idx := range cp.Controllers {
		if cp.Controllers[idx].Name == name {
			ret = &cp.Controllers[idx]
			return
		}
	}
	err = util.NewError("Could not find Controller " + name)
	return
}

func (cp *RemoteControlPlane) GetEndpoint() (string, error) {
	if cp.Endpoint != "" {
		return cp.Endpoint, nil
	}
	if cp.Controller.PublicUrl != "" {
		return cp.Controller.PublicUrl, nil
	}
	if len(cp.Controllers) == 0 {
		return "", util.NewInternalError("Control Plane does not have any Controllers")
	}
	for idx := range cp.Controllers {
		if cp.Controllers[idx].Endpoint != "" {
			return cp.Controllers[idx].Endpoint, nil
		}
	}
	return "", util.NewInternalError("No Controllers in Remote Control Plane had an endpoint available")
}

func (cp *RemoteControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*RemoteController)
	if !ok {
		return util.NewError("Must add Remote Controller to Remote Control Plane")
	}
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == controller.GetName() {
			cp.Controllers[idx] = *controller
			return nil
		}
	}
	cp.Controllers = append(cp.Controllers, *controller)
	return nil
}

func (cp *RemoteControlPlane) AddController(baseController Controller) error {
	if _, err := cp.GetController(baseController.GetName()); err == nil {
		return util.NewError("Could not add Controller " + baseController.GetName() + " because it already exists")
	}
	controller, ok := baseController.(*RemoteController)
	if !ok {
		return util.NewError("Must add Remote Controller to Remote Control Plane")
	}

	cp.Controllers = append(cp.Controllers, *controller)
	return nil
}

func (cp *RemoteControlPlane) DeleteController(name string) error {
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == name {
			cp.Controllers = append(cp.Controllers[:idx], cp.Controllers[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name + " when performing deletion")
}

func (cp *RemoteControlPlane) Sanitize() (err error) {
	for idx := range cp.Controllers {
		if err = cp.Controllers[idx].Sanitize(); err != nil {
			return
		}
	}
	return nil
}

const controllerAddOnConnectHint = "use connect -f with a full controlplane.yaml or deploy -f controlplane.yaml first"

// SupportsControllerAddOn reports whether the stored Control Plane has enough deploy
// metadata to add controllers via standalone kind: Controller YAML.
func (cp *RemoteControlPlane) SupportsControllerAddOn() error {
	if cp.Auth.Mode == "" {
		return util.NewInputError("namespace Control Plane does not support adding controllers; " + controllerAddOnConnectHint)
	}
	if !hasControllerSpec(cp.Controller) {
		return util.NewInputError("namespace Control Plane is missing spec.controller; " + controllerAddOnConnectHint)
	}
	return nil
}

// ValidateControllerAddOn rejects duplicate controller name or host in the stored CP.
func (cp *RemoteControlPlane) ValidateControllerAddOn(ctrl *RemoteController) error {
	for _, existing := range cp.Controllers {
		if existing.Name == ctrl.Name {
			return util.NewInputError(fmt.Sprintf("controller name %q already exists in namespace Control Plane", ctrl.Name))
		}
		if existing.Host != "" && existing.Host == ctrl.Host {
			return util.NewInputError(fmt.Sprintf("controller host %q already exists in namespace Control Plane", ctrl.Host))
		}
	}
	return nil
}

// ValidateControllerAddOnDatabase rejects SQLite when adding a second controller.
func (cp *RemoteControlPlane) ValidateControllerAddOnDatabase() error {
	if len(cp.Controllers) >= 1 && cp.Database.Provider == "" {
		return util.NewInputError("cannot add controller: external database is required when multiple controllers are configured")
	}
	return nil
}

func hasControllerSpec(c LocalControllerSpec) bool {
	if c.PublicUrl != "" || c.ConsoleUrl != "" || c.LogLevel != "" || c.PidBaseDir != "" || c.Package != nil {
		return true
	}
	if c.ConsolePort != 0 || c.TrustProxy != nil || c.Https != nil || c.SecretName != "" {
		return true
	}
	return false
}

func (cp *RemoteControlPlane) Clone() ControlPlane {
	controllers := make([]RemoteController, len(cp.Controllers))
	copy(controllers, cp.Controllers)
	return &RemoteControlPlane{
		Endpoint:            cp.Endpoint,
		CA:                  cp.CA,
		IofogUser:           cp.IofogUser,
		Controller:          cp.Controller,
		Controllers:         controllers,
		Database:            cp.Database,
		Auth:                cp.Auth,
		RouterSiteCA:        cp.RouterSiteCA,
		RouterLocalCA:       cp.RouterLocalCA,
		NatsSiteCA:          cp.NatsSiteCA,
		NatsLocalCA:         cp.NatsLocalCA,
		SystemMicroservices: cp.SystemMicroservices,
		Nats:                cp.Nats,
		Events:              cp.Events,
		Vault:               cp.Vault,
		TLS:                 cp.TLS,
		Airgap:              cp.Airgap,
	}
}
