package resource

import (
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type KubernetesControlPlane struct {
	KubeConfig     string                 `yaml:"config"`
	IofogUser      IofogUser              `yaml:"iofogUser"`
	ControllerPods []KubernetesController `yaml:"controllerPods,omitempty"`
	Database       Database               `yaml:"database"`
	Auth           Auth                   `yaml:"auth"`
	Events         Events                 `yaml:"events,omitempty"`
	Services       Services               `yaml:"services,omitempty"`
	Replicas       Replicas               `yaml:"replicas,omitempty"`
	Images         KubeImages             `yaml:"images,omitempty"`
	Endpoint       string                 `yaml:"endpoint,omitempty"`
	Controller     K8SControllerConfig    `yaml:"controller,omitempty"`
	Ingresses      Ingresses              `yaml:"ingresses,omitempty"`
	Nats           *NatsSpec              `yaml:"nats,omitempty"`
	Vault          *VaultSpec             `yaml:"vault,omitempty"`
}

func (cp *KubernetesControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp *KubernetesControlPlane) UpdateUserTokens(accessToken, refreshToken string) IofogUser {
	cp.IofogUser.AccessToken = accessToken
	cp.IofogUser.RefreshToken = refreshToken

	return cp.IofogUser
}

func (cp *KubernetesControlPlane) GetControllers() (controllers []Controller) {
	for idx := range cp.ControllerPods {
		controllers = append(controllers, cp.ControllerPods[idx].Clone())
	}
	return
}

func (cp *KubernetesControlPlane) GetController(name string) (ret Controller, err error) {
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == name {
			ret = &cp.ControllerPods[idx]
			return
		}
	}
	err = util.NewError("Could not find Controller " + name)
	return
}

func (cp *KubernetesControlPlane) GetEndpoint() (string, error) {
	// If HTTPS is enabled in the controller configuration, regenerate the endpoint with HTTPS
	if cp.Controller.Https != nil && *cp.Controller.Https {
		if cp.Endpoint != "" {
			// Endpoint already uses HTTPS (e.g. from ingress: https://hostname); return as-is
			// to avoid util.GetControllerEndpoint incorrectly adding :51121 when given a hostname without scheme
			if strings.HasPrefix(cp.Endpoint, "https://") {
				return cp.Endpoint, nil
			}
			// Convert http to https
			host := cp.Endpoint
			if len(host) > 7 && host[:7] == "http://" {
				host = host[7:]
			}
			return util.GetControllerEndpoint(host, true)
		}
	}
	return cp.Endpoint, nil
}

func (cp *KubernetesControlPlane) UpdateController(baseController Controller) error {
	controller, ok := baseController.(*KubernetesController)
	if !ok {
		return util.NewError("Must add Kubernetes Controller to Kubernetes Control Plane")
	}
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == controller.GetName() {
			cp.ControllerPods[idx] = *controller
			return nil
		}
	}
	cp.ControllerPods = append(cp.ControllerPods, *controller)
	return nil
}

func (cp *KubernetesControlPlane) AddController(baseController Controller) error {
	if _, err := cp.GetController(baseController.GetName()); err == nil {
		return util.NewError("Could not add Controller " + baseController.GetName() + " because it already exists")
	}
	controller, ok := baseController.(*KubernetesController)
	if !ok {
		return util.NewError("Must add Kubernetes Controller to Kubernetes Control Plane")
	}
	cp.ControllerPods = append(cp.ControllerPods, *controller)
	return nil
}

func (cp *KubernetesControlPlane) DeleteController(name string) error {
	for idx := range cp.ControllerPods {
		if cp.ControllerPods[idx].GetName() == name {
			cp.ControllerPods = append(cp.ControllerPods[:idx], cp.ControllerPods[idx+1:]...)
			return nil
		}
	}
	return util.NewError("Could not find Controller " + name + " when performing deletion")
}

func (cp *KubernetesControlPlane) Sanitize() (err error) {
	if cp.KubeConfig, err = util.FormatPath(cp.KubeConfig); err != nil {
		return
	}
	if cp.Replicas.Controller == 0 {
		cp.Replicas.Controller = 1
	}
	return
}

func (cp *KubernetesControlPlane) ValidateKubeConfig() error {
	if cp.KubeConfig == "" {
		return NewNoKubeConfigError("Control Plane")
	}
	return nil
}

func (cp *KubernetesControlPlane) Clone() ControlPlane {
	controllerPods := make([]KubernetesController, len(cp.ControllerPods))
	copy(controllerPods, cp.ControllerPods)
	return &KubernetesControlPlane{
		KubeConfig:     cp.KubeConfig,
		IofogUser:      cp.IofogUser,
		Auth:           cp.Auth,
		Database:       cp.Database,
		Events:         cp.Events,
		Services:       cp.Services,
		Ingresses:      cp.Ingresses,
		Replicas:       cp.Replicas,
		Images:         cp.Images,
		Endpoint:       cp.Endpoint,
		Controller:     cp.Controller,
		ControllerPods: controllerPods,
		Nats:           cp.Nats,
		Vault:          cp.Vault,
	}
}
