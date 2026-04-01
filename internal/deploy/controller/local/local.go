/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package deploylocalcontroller

import (
	"fmt"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace             string
	ctrl                  *rsc.LocalController
	ctrlPlane             *rsc.LocalControlPlane
	client                *install.LocalContainer
	localControllerConfig *install.LocalContainerConfig
	containersNames       []string
	iofogUser             rsc.IofogUser
}

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	controller, err := rsc.UnmarshallLocalController(opt.Yaml)
	if err != nil {
		return
	}

	if len(opt.Name) > 0 {
		controller.Name = opt.Name
	}

	// Validate
	if err = Validate(&controller); err != nil {
		return
	}

	// Get the Control Plane
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	// controlPlane, err := ns.GetControlPlane()
	// if err != nil {
	// 	return
	// }

	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return
	}
	controlPlane, ok := baseControlPlane.(*rsc.LocalControlPlane)
	if !ok {
		err = util.NewError("Could not convert Control Plane to Remote Control Plane")
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, controlPlane, &controller)
}

func NewExecutorWithoutParsing(namespace string, controlPlane *rsc.LocalControlPlane, controller *rsc.LocalController) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}
	if err := util.IsLowerAlphanumeric("Controller", controller.GetName()); err != nil {
		return nil, err
	}
	cli, err := install.NewLocalContainerClient()
	if err != nil {
		return nil, err
	}

	// Instantiate executor
	return newExecutor(namespace, controlPlane, controller, cli), nil
}

// TODO: Rewrite this pkg, don't need ctrl coming in here
func newExecutor(namespace string, controlPlane *rsc.LocalControlPlane, ctrl *rsc.LocalController, client *install.LocalContainer) *localExecutor {
	return &localExecutor{
		namespace: namespace,
		ctrl:      ctrl,
		client:    client,
		localControllerConfig: install.NewLocalControllerConfig(ctrl.Container.Image, install.Credentials{
			User:     ctrl.Container.Credentials.User,
			Password: ctrl.Container.Credentials.Password,
		}, install.Auth{
			URL:              controlPlane.Auth.URL,
			Realm:            controlPlane.Auth.Realm,
			SSL:              controlPlane.Auth.SSL,
			RealmKey:         controlPlane.Auth.RealmKey,
			ControllerClient: controlPlane.Auth.ControllerClient,
			ControllerSecret: controlPlane.Auth.ControllerSecret,
			ViewerClient:     controlPlane.Auth.ViewerClient,
		}, install.Database{
			Provider:     controlPlane.Database.Provider,
			Host:         controlPlane.Database.Host,
			Port:         controlPlane.Database.Port,
			User:         controlPlane.Database.User,
			Password:     controlPlane.Database.Password,
			DatabaseName: controlPlane.Database.DatabaseName,
			SSL:          controlPlane.Database.SSL,
			CA:           controlPlane.Database.CA,
		}, install.Events{
			AuditEnabled:     controlPlane.Events.AuditEnabled,
			RetentionDays:    controlPlane.Events.RetentionDays,
			CleanupInterval:  controlPlane.Events.CleanupInterval,
			CaptureIpAddress: controlPlane.Events.CaptureIpAddress,
		}, localSystemImagesToInstall(controlPlane.SystemMicroservices, controlPlane.Nats)),
		iofogUser: controlPlane.GetUser(),
		ctrlPlane: controlPlane,
	}
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", errClean))
		}
	}
}

func (exe *localExecutor) deployContainers() error {
	controllerContainerConfig := exe.localControllerConfig
	controllerContainerName := controllerContainerConfig.ContainerName

	// Deploy controller image
	util.SpinStart("Deploying Controller container")

	// If container already exists, clean it
	if _, err := exe.client.GetContainerByName(controllerContainerName); err == nil {
		if err := exe.client.CleanContainer(controllerContainerName); err != nil {
			return err
		}
	}

	_, err := exe.client.DeployContainer(controllerContainerConfig)
	if err != nil {
		return err
	}

	exe.containersNames = append(exe.containersNames, controllerContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Controller API")
	if err := exe.client.WaitForCommand(
		install.GetLocalContainerName("controller", false),
		regexp.MustCompile("\"status\":[ |\t]*\"online\""),
		"iofog-controller",
		"controller",
		"status",
	); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *localExecutor) Execute() error {
	// Deploy Controller images
	if err := exe.deployContainers(); err != nil {
		exe.cleanContainers()
		return err
	}

	// Update controller (its a pointer, this is returned to caller)
	controllerContainerConfig := exe.localControllerConfig
	// Local controllers typically use HTTP by default
	endpoint, err := util.GetControllerEndpoint(fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host), false)
	if err != nil {
		return err
	}

	exe.ctrl.Endpoint = endpoint
	exe.ctrl.Created = util.NowUTC()
	return exe.ctrlPlane.UpdateController(exe.ctrl)
}

func localSystemImagesToInstall(s *rsc.LocalSystemMicroservices, nats *rsc.NatsEnabledConfig) *install.LocalSystemImages {
	// Always pass a non-nil struct so local controller gets default NATS_IMAGE_1/2 and NATS_ENABLED when not overridden
	out := &install.LocalSystemImages{}
	if s != nil {
		out.Router = s.Router
		out.Nats = s.Nats
	}
	if nats != nil && nats.Enabled != nil {
		out.NatsEnabled = nats.Enabled
	}
	return out
}

func Validate(ctrl rsc.Controller) error {
	if err := util.IsLowerAlphanumeric("Controller", ctrl.GetName()); err != nil {
		return err
	}
	return nil
}
