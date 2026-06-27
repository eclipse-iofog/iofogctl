package resource

import (
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type LocalControlPlane struct {
	Endpoint            string                            `yaml:"endpoint,omitempty"`
	CA                  string                            `yaml:"ca,omitempty"`
	IofogUser           IofogUser                         `yaml:"iofogUser"`
	Controller          LocalControllerSpec               `yaml:"controller,omitempty"`
	Controllers         []LocalController                 `yaml:"controllers,omitempty"`
	Database            Database                          `yaml:"database"`
	Auth                Auth                              `yaml:"auth"`
	RouterSiteCA        *SiteCertificate                  `yaml:"routerSiteCA,omitempty"`
	RouterLocalCA       *SiteCertificate                  `yaml:"routerLocalCA,omitempty"`
	NatsSiteCA          *SiteCertificate                  `yaml:"natsSiteCA,omitempty"`
	NatsLocalCA         *SiteCertificate                  `yaml:"natsLocalCA,omitempty"`
	SystemMicroservices install.RemoteSystemMicroservices `yaml:"systemMicroservices,omitempty"`
	Nats                *NatsEnabledConfig                `yaml:"nats,omitempty"`
	Events              Events                            `yaml:"events,omitempty"`
	Vault               *VaultSpec                        `yaml:"vault,omitempty"`
	TLS                 *ControlPlaneTLS                  `yaml:"tls,omitempty"`
	SystemAgent         *SystemAgentConfig                `yaml:"systemAgent,omitempty"`
	Airgap              bool                              `yaml:"airgap,omitempty"`
}

func (cp *LocalControlPlane) GetTrustCA() string {
	return cp.CA
}

func (cp *LocalControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp *LocalControlPlane) UpdateUserTokens(accessToken, refreshToken string) IofogUser {
	cp.IofogUser.AccessToken = accessToken
	cp.IofogUser.RefreshToken = refreshToken

	return cp.IofogUser
}

func (cp *LocalControlPlane) GetControllers() []Controller {
	controllers := make([]Controller, 0, len(cp.Controllers))
	for idx := range cp.Controllers {
		controllers = append(controllers, cp.Controllers[idx].Clone())
	}
	return controllers
}

func (cp *LocalControlPlane) GetController(name string) (Controller, error) {
	for idx := range cp.Controllers {
		if name == "" || cp.Controllers[idx].GetName() == name {
			return &cp.Controllers[idx], nil
		}
	}
	return nil, util.NewError("Local Control Plane does not have a Controller")
}

func (cp *LocalControlPlane) GetEndpoint() (string, error) {
	if cp.Endpoint != "" {
		return cp.Endpoint, nil
	}
	if cp.Controller.PublicUrl != "" {
		return cp.Controller.PublicUrl, nil
	}
	for idx := range cp.Controllers {
		if cp.Controllers[idx].Endpoint != "" {
			return cp.Controllers[idx].Endpoint, nil
		}
	}
	return "", util.NewError("Local Control Plane does not have an endpoint")
}

func (cp *LocalControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
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

func (cp *LocalControlPlane) AddController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
	}
	for idx := range cp.Controllers {
		if cp.Controllers[idx].GetName() == controller.GetName() {
			return util.NewConflictError(controller.GetName())
		}
	}
	cp.Controllers = append(cp.Controllers, *controller)
	return nil
}

func (cp *LocalControlPlane) DeleteController(name string) error {
	for idx := range cp.Controllers {
		if name == "" || cp.Controllers[idx].GetName() == name {
			cp.Controllers = append(cp.Controllers[:idx], cp.Controllers[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name)
}

func (cp *LocalControlPlane) Sanitize() error {
	for idx := range cp.Controllers {
		if err := cp.Controllers[idx].Sanitize(); err != nil {
			return err
		}
		if !util.IsLocalHost(cp.Controllers[idx].Endpoint) {
			cp.Controllers[idx].Endpoint = "localhost"
		}
	}
	return nil
}

func (cp *LocalControlPlane) Clone() ControlPlane {
	controllers := make([]LocalController, len(cp.Controllers))
	copy(controllers, cp.Controllers)
	var systemAgent *SystemAgentConfig
	if cp.SystemAgent != nil {
		systemAgent = &SystemAgentConfig{}
		*systemAgent = *cp.SystemAgent
	}
	return &LocalControlPlane{
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
		SystemAgent:         systemAgent,
		Airgap:              cp.Airgap,
	}
}
