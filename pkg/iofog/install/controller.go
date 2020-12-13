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

package install

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type ControllerOptions struct {
	User                string
	Host                string
	Port                int
	PrivKeyFilename     string
	Version             string
	Repo                string
	Token               string
	SystemMicroservices rsc.RemoteSystemMicroservices
	PidBaseDir          string
	EcnViewerPort       int
}

type database struct {
	databaseName string
	provider     string
	host         string
	user         string
	password     string
	port         int
}

type Controller struct {
	*ControllerOptions
	ssh *util.SecureShellClient
	db  database
}

func NewController(options *ControllerOptions) *Controller {
	ssh := util.NewSecureShellClient(options.User, options.Host, options.PrivKeyFilename)
	ssh.SetPort(options.Port)
	if options.Version == "" || options.Version == "latest" {
		options.Version = util.GetControllerVersion()
	}
	return &Controller{
		ControllerOptions: options,
		ssh:               ssh,
	}
}

func (ctrl *Controller) SetControllerExternalDatabase(host, user, password, provider, databaseName string, port int) {
	if provider == "" {
		provider = "postgres"
	}
	if databaseName == "" {
		databaseName = "iofogcontroller"
	}
	ctrl.db = database{
		databaseName: databaseName,
		provider:     provider,
		host:         host,
		user:         user,
		password:     password,
		port:         port,
	}
}

func (ctrl *Controller) CopyScript(srcDir, filename, destDir string) (err error) {
	// Read script from assets
	if srcDir != "" {
		srcDir = util.AddTrailingSlash(srcDir)
	}
	staticFile := util.GetStaticFile(srcDir + filename)

	// Copy to /tmp for backwards compatability
	for _, currDestDir := range []string{util.AddTrailingSlash(destDir), "/tmp/"} {
		reader := strings.NewReader(staticFile)
		if err := ctrl.ssh.CopyTo(reader, currDestDir, filename, "0775", int64(len(staticFile))); err != nil {
			return err
		}
	}

	return nil
}

func (ctrl *Controller) Uninstall() (err error) {
	// Stop controller gracefully
	if err = ctrl.Stop(); err != nil {
		return err
	}

	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// Copy uninstallation scripts to remote host
	Verbose("Copying install files to server")
	scripts := []string{
		"controller_uninstall_iofog.sh",
	}
	for _, script := range scripts {
		if err = ctrl.CopyScript("", script); err != nil {
			return err
		}
	}

	cmds := []command{
		{
			cmd: "sudo /tmp/controller_uninstall_iofog.sh",
			msg: "Uninstalling controller on host " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err = ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return
		}
	}
	return nil
}

func (ctrl *Controller) Install() (err error) {
	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// Copy installation scripts to remote host
	Verbose("Copying install files to server")
	scripts := []string{
		"check_prereqs.sh",
		"controller_install_node.sh",
		"controller_install_iofog.sh",
		"controller_set_env.sh",
	}
	for _, script := range scripts {
		if err = ctrl.CopyScript("", script); err != nil {
			return err
		}
	}

	// Copy service scripts to remote host
	Verbose("Copying service files to server")
	if _, err = ctrl.ssh.Run("mkdir -p /tmp/iofog-controller-service"); err != nil {
		return err
	}
	scripts = []string{
		"iofog-controller.initctl",
		"iofog-controller.systemd",
		"iofog-controller.update-rc",
	}
	for _, script := range scripts {
		if err = ctrl.CopyScript("iofog-controller-service/", script); err != nil {
			return err
		}
	}

	// Encode environment variables
	env := make([]string, 0)
	if ctrl.db.host != "" {
		env = append(env, fmt.Sprintf("\"DB_PROVIDER=%s\"", ctrl.db.provider))
		env = append(env, fmt.Sprintf("\"DB_HOST=%s\"", ctrl.db.host))
		env = append(env, fmt.Sprintf("\"DB_USER=%s\"", ctrl.db.user))
		env = append(env, fmt.Sprintf("\"DB_PASSWORD=%s\"", ctrl.db.password))
		env = append(env, fmt.Sprintf("\"DB_PORT=%d\"", ctrl.db.port))
		env = append(env, fmt.Sprintf("\"DB_NAME=%s\"", ctrl.db.databaseName))
	}
	if ctrl.PidBaseDir != "" {
		env = append(env, fmt.Sprintf("\"PID_BASE=%s\"", ctrl.PidBaseDir))
	}
	if ctrl.EcnViewerPort != 0 {
		env = append(env, fmt.Sprintf("\"VIEWER_PORT=%d\"", ctrl.EcnViewerPort))
	}
	if ctrl.SystemMicroservices.Proxy.X86 != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Proxy_1=%s\"", ctrl.SystemMicroservices.Proxy.X86))
	}
	if ctrl.SystemMicroservices.Proxy.ARM != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Proxy_2=%s\"", ctrl.SystemMicroservices.Proxy.ARM))
	}
	if ctrl.SystemMicroservices.Router.X86 != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Router_1=%s\"", ctrl.SystemMicroservices.Router.X86))
	}
	if ctrl.SystemMicroservices.Router.ARM != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Router_2=%s\"", ctrl.SystemMicroservices.Router.ARM))
	}

	envString := strings.Join(env, " ")

	// Define commands
	cmds := []command{
		{
			cmd: "/tmp/check_prereqs.sh",
			msg: "Checking prerequisites on Controller " + ctrl.Host,
		},
		{
			cmd: "sudo /tmp/controller_install_node.sh",
			msg: "Installing Node.js on Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo /tmp/controller_set_env.sh %s", envString),
			msg: "Setting up environment variables for Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo /tmp/controller_install_iofog.sh %s %s %s", ctrl.Version, ctrl.Repo, ctrl.Token),
			msg: "Installing ioFog on Controller " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err = ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return
		}
	}

	// Specify errors to ignore while waiting
	ignoredErrors := []string{
		"Process exited with status 7", // curl: (7) Failed to connect to localhost port 8080: Connection refused
	}
	// Wait for Controller
	Verbose("Waiting for Controller " + ctrl.Host)
	if err = ctrl.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"online\""),
		fmt.Sprintf("curl --request GET --url http://localhost:%s/api/v3/status", iofog.ControllerPortString),
		ignoredErrors,
	); err != nil {
		return
	}

	// Wait for API
	endpoint := fmt.Sprintf("%s:%s", ctrl.Host, iofog.ControllerPortString)
	if err = WaitForControllerAPI(endpoint); err != nil {
		return
	}

	return
}

func (ctrl *Controller) Stop() (err error) {
	// Connect to server
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// TODO: Clear the database
	// Define commands
	cmds := []string{
		"sudo iofog-controller stop",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}

func WaitForControllerAPI(endpoint string) (err error) {
	ctrlClient := client.New(client.Options{Endpoint: endpoint})

	seconds := 0
	for seconds < 60 {
		// Try to create the user, return if success
		if _, err = ctrlClient.GetStatus(); err == nil {
			return
		}
		// Connection failed, wait and retry
		time.Sleep(time.Millisecond * 1000)
		seconds = seconds + 1
	}

	// Return last error
	return
}
