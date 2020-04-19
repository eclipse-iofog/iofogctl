/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package resource

import (
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type RemoteSystemImages struct {
	ARM string `yaml:"arm,omitempty"`
	X86 string `yaml:"x86,omitempty"`
}

type RemoteSystemMicroservices struct {
	Router RemoteSystemImages `yaml:"router,omitempty"`
	Proxy  RemoteSystemImages `yaml:"proxy,omitempty"`
}

type RemoteControlPlane struct {
	IofogUser           IofogUser                 `yaml:"iofogUser"`
	Controllers         []RemoteController        `yaml:"controllers"`
	Database            Database                  `yaml:"database,omitempty"`
	Package             Package                   `yaml:"package,omitempty"`
	SystemAgent         Package                   `yaml:"systemAgent,omitempty"`
	SystemMicroservices RemoteSystemMicroservices `yaml:"systemMicroservices,omitempty"`
}

func (cp RemoteControlPlane) GetUser() IofogUser {
	return cp.IofogUser
}

func (cp RemoteControlPlane) GetControllers() (controllers []Controller) {
	for idx := range cp.Controllers {
		controllers = append(controllers, cp.Controllers[idx].Clone())
	}
	return
}

func (cp RemoteControlPlane) GetController(name string) (ret Controller, err error) {
	for _, ctrl := range cp.Controllers {
		if ctrl.GetName() == name {
			ret = &ctrl
			return
		}
	}
	err = util.NewError("Could not find Controller " + name)
	return
}

func (cp RemoteControlPlane) GetEndpoint() (string, error) {
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
	cp.IofogUser.EncodePassword()
	for idx := range cp.Controllers {
		if err = cp.Controllers[idx].Sanitize(); err != nil {
			return
		}
	}
	return nil
}

func (cp *RemoteControlPlane) Clone() ControlPlane {
	controllers := make([]RemoteController, len(cp.Controllers))
	copy(controllers, cp.Controllers)
	return &RemoteControlPlane{
		IofogUser:           cp.IofogUser,
		Database:            cp.Database,
		Package:             cp.Package,
		SystemAgent:         cp.SystemAgent,
		SystemMicroservices: cp.SystemMicroservices,
		Controllers:         controllers,
	}
}
