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

package deploylocalcontrolplane

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"

	// deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agentconfig"
	deploylocalcontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller/local"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"

	// clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}
type localControlPlaneExecutor struct {
	ctrlClient          *client.Client
	controllerExecutors []execute.Executor
	controlPlane        *rsc.LocalControlPlane
	namespace           string
	name                string
}

// func createDefaultRouter(clt *client.Client) (err error) {
// 	routerConfig := client.Router{
// 		Host: "localhost",
// 		RouterConfig: client.RouterConfig{
// 			RouterMode:      iutil.MakeStrPtr("interior"),
// 			MessagingPort:   iutil.MakeIntPtr(5671),
// 			EdgeRouterPort:  iutil.MakeIntPtr(45671),
// 			InterRouterPort: iutil.MakeIntPtr(55671),
// 		},
// 	}

// 	return clt.PutDefaultRouter(routerConfig)
// }

// prepareViewerURL prepares the viewer URL from endpoint using logic similar to view.go
func prepareViewerURL(endpoint string) (string, error) {
	URL, err := url.Parse(endpoint)
	if err != nil || URL.Host == "" {
		URL, err = url.Parse("//" + endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse endpoint: %v", err)
		}
	}

	if URL.Scheme == "" {
		URL.Scheme = "http"
	}

	host := ""
	if strings.Contains(URL.Host, ":") {
		host, _, err = net.SplitHostPort(URL.Host)
		if err != nil {
			return "", fmt.Errorf("failed to split host and port: %v", err)
		}
	} else {
		host = URL.Host
	}

	// Add port for localhost
	if util.IsLocalHost(host) {
		host = net.JoinHostPort(host, iofog.ControllerHostECNViewerPortString)
	}

	URL.Host = host
	return URL.String(), nil
}

// updateViewerClientRootURL updates the viewer client root URL in Keycloak if auth is configured
func updateViewerClientRootURL(controlPlane *rsc.LocalControlPlane, endpoint string) error {
	// Check if auth is configured
	if controlPlane.Auth.URL == "" || controlPlane.Auth.ViewerClient == "" {
		// Auth not configured, skip update
		return nil
	}

	// Prepare viewer URL
	viewerURL, err := prepareViewerURL(endpoint)
	if err != nil {
		return fmt.Errorf("failed to prepare viewer URL: %v", err)
	}

	// Update viewer client root URL
	if err := iutil.UpdateECNViewerClientRootURL(controlPlane.Auth, viewerURL); err != nil {
		return fmt.Errorf("failed to update viewer client root URL: %v", err)
	}

	return nil
}

func (exe localControlPlaneExecutor) postDeploy() (err error) {
	// Check controller is reachable
	// clt, err := clientutil.NewControllerClient(exe.namespace)
	// if err != nil {
	// 	return err
	// }

	// if err := createDefaultRouter(clt); err != nil {
	// 	return err
	// }
	return nil
}

func (exe localControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))
	if err := runExecutors(exe.controllerExecutors); err != nil {
		return err
	}

	// Make sure Controller API is ready
	controller, err := exe.controlPlane.GetController("")
	if err != nil {
		return err
	}
	endpoint := controller.GetEndpoint()

	if err := install.WaitForControllerAPI(endpoint); err != nil {
		return err
	}

	// // Create new user
	// baseURL, err := util.GetBaseURL(endpoint)
	// if err != nil {
	// 	return err
	// }
	// exe.ctrlClient = client.New(client.Options{BaseURL: baseURL})
	// user := client.User(exe.controlPlane.GetUser())
	// user.Password = exe.controlPlane.GetUser().GetRawPassword()
	// if err = exe.ctrlClient.CreateUser(user); err != nil {
	// 	// If not error about account existing, fail
	// 	if !strings.Contains(err.Error(), "already an account associated") {
	// 		return err
	// 	}
	// 	// Try to log in
	// 	if err := exe.ctrlClient.Login(client.LoginRequest{
	// 		Email:    user.Email,
	// 		Password: user.Password,
	// 	}); err != nil {
	// 		return err
	// 	}
	// }
	// Update config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	if err := config.Flush(); err != nil {
		return err
	}

	// Update viewer client root URL if auth is configured
	if err := updateViewerClientRootURL(exe.controlPlane, endpoint); err != nil {
		// Log error but don't fail deployment
		util.PrintInfo(fmt.Sprintf("Warning: Failed to update viewer client root URL: %v\n", err))
	}

	// Post deploy steps
	return exe.postDeploy()
}

func (exe localControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(executors []execute.Executor, namespace, name string, controlPlane *rsc.LocalControlPlane) execute.Executor {
	return localControlPlaneExecutor{
		controllerExecutors: executors,
		namespace:           namespace,
		controlPlane:        controlPlane,
		name:                name,
	}
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	_, err = config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	// Read the input file
	controlPlane, err := rsc.UnmarshallLocalControlPlane(opt.Yaml)
	if err != nil {
		return
	}

	// Create exe Controllers
	controllers := controlPlane.GetControllers()
	controllerExecutors := make([]execute.Executor, len(controllers))
	for idx := range controllers {
		controller, ok := controllers[idx].(*rsc.LocalController)
		if !ok {
			return nil, util.NewError("Could not convert Controller to Local Controller")
		}
		exe, err := deploylocalcontroller.NewExecutorWithoutParsing(opt.Namespace, &controlPlane, controller)
		if err != nil {
			return nil, err
		}
		controllerExecutors[idx] = exe
	}

	return newControlPlaneExecutor(controllerExecutors, opt.Namespace, opt.Name, &controlPlane), nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}
