/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package connectremotecontrolplane

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
}

func NewManualExecutor(namespace, name, endpoint, email, password string) (execute.Executor, error) {
	controlPlane := &rsc.RemoteControlPlane{
		IofogUser: rsc.IofogUser{
			Email:    email,
			Password: password,
		},
		// TODO: This is wrong, could be a list of controllers in a control plane
		Controllers: []rsc.RemoteController{
			{
				Name:     name,
				Endpoint: formatEndpoint(endpoint),
			},
		},
	}

	return newRemoteExecutor(controlPlane, namespace), nil
}

func NewExecutor(namespace, name string, yaml []byte, kind config.Kind) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := rsc.UnmarshallRemoteControlPlane(yaml)
	if err != nil {
		return nil, err
	}

	if err := validate(&controlPlane); err != nil {
		return nil, err
	}

	// In YAML, the endpoint will come through Host variable
	for _, baseController := range controlPlane.GetControllers() {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return nil, util.NewError("Could not convert Controller to Remote Controller")
		}
		controller.Endpoint = formatEndpoint(controlPlane.Controllers[0].Host)
		if err := controlPlane.UpdateController(controller); err != nil {
			return nil, err
		}
	}

	return newRemoteExecutor(&controlPlane, namespace), nil
}

func newRemoteExecutor(controlPlane *rsc.RemoteControlPlane, namespace string) *remoteExecutor {
	r := &remoteExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
	}
	return r
}

func (exe *remoteExecutor) GetName() string {
	return "Remote Control Plane"
}

func (exe *remoteExecutor) Execute() (err error) {
	// Establish connection
	controllers := exe.controlPlane.GetControllers()
	if len(controllers) == 0 {
		return util.NewError("Control Plane in Namespace " + exe.namespace + " has no Controllers. Try deploying a Control Plane to this Namespace.")
	}
	endpoint, err := exe.controlPlane.GetEndpoint()
	if err != nil {
		return err
	}
	err = connect(exe.controlPlane, endpoint, exe.namespace)
	if err != nil {
		return err
	}
	// Save result
	config.UpdateControlPlane(exe.namespace, exe.controlPlane)
	return config.Flush()
}

// TODO: remove duplication
func connect(ctrlPlane rsc.ControlPlane, endpoint, namespace string) error {
	// Connect to Controller
	ctrl, err := client.NewAndLogin(client.Options{Endpoint: endpoint}, ctrlPlane.GetUser().Email, ctrlPlane.GetUser().Password)
	if err != nil {
		return err
	}

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := rsc.RemoteAgent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err = config.AddAgent(namespace, &agentConfig); err != nil {
			return err
		}
	}

	return nil
}

// TODO: remove duplication
func formatEndpoint(endpoint string) string {
	before := util.Before(endpoint, ":")
	after := util.After(endpoint, ":")
	if after == "" {
		after = iofog.ControllerPortString
	}
	return before + ":" + after
}

func validate(controlPlane rsc.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
	if user.Password == "" || user.Email == "" {
		return util.NewInputError("To connect, Control Plane Iofog User must contain non-empty values in email and password fields")
	}
	// Validate Controllers
	if len(controlPlane.GetControllers()) == 0 {
		err = util.NewInputError("Control Plane must have at least one Controller instance specified.")
		return
	}
	for _, ctrl := range controlPlane.GetControllers() {
		if ctrl.GetName() == "" {
			return util.NewInputError("You must specify a non-empty value for name value of Controllers")
		}
	}

	return
}
