package resource

import (
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type LocalControlPlane struct {
	IofogUser           IofogUser                 `yaml:"iofogUser"`
	Controller          *LocalController          `yaml:"controller,omitempty"`
	Database            Database                  `yaml:"database"`
	Auth                Auth                      `yaml:"auth"`
	Events              Events                    `yaml:"events,omitempty"`
	SystemMicroservices *LocalSystemMicroservices `yaml:"systemMicroservices,omitempty"`
	Nats                *NatsEnabledConfig        `yaml:"nats,omitempty"`
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
	if cp.Controller == nil {
		return []Controller{}
	}
	return []Controller{cp.Controller.Clone()}
}

func (cp *LocalControlPlane) GetController(name string) (Controller, error) {
	if cp.Controller == nil {
		return nil, util.NewError("Local Control Plane does not have a Controller")
	}
	return cp.Controller, nil
}

func (cp *LocalControlPlane) GetEndpoint() (string, error) {
	if cp.Controller == nil {
		return "", util.NewError("Local Control Plane does not have a Controller, cannot get endpoint.")
	}
	return cp.Controller.GetEndpoint(), nil
}

func (cp *LocalControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) AddController(baseController Controller) error {
	controller, ok := baseController.(*LocalController)
	if !ok {
		return util.NewError("Must add Local Controller to Local Control Plane")
	}
	cp.Controller = controller
	return nil
}

func (cp *LocalControlPlane) DeleteController(string) error {
	cp.Controller = nil
	return nil
}

func (cp *LocalControlPlane) Sanitize() error {
	if cp.Controller != nil && !util.IsLocalHost(cp.Controller.Endpoint) {
		cp.Controller.Endpoint = "localhost"
	}
	return nil
}

func (cp *LocalControlPlane) Clone() ControlPlane {
	var sys *LocalSystemMicroservices
	if cp.SystemMicroservices != nil {
		sys = &LocalSystemMicroservices{
			Router: cp.SystemMicroservices.Router,
			Nats:   cp.SystemMicroservices.Nats,
		}
	}
	return &LocalControlPlane{
		IofogUser:           cp.IofogUser,
		Controller:          cp.Controller.Clone().(*LocalController),
		Database:            cp.Database,
		Auth:                cp.Auth,
		Events:              cp.Events,
		SystemMicroservices: sys,
		Nats:                cp.Nats,
	}
}
