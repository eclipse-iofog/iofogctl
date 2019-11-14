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

package install

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type ControllerOptions struct {
	User            string
	Host            string
	Port            int
	PrivKeyFilename string
	Version         string
	Repo            string
	Token           string
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
		options.Version = util.GetControllerTag()
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

func (ctrl *Controller) CopyScript(path string, name string) (err error) {
	script := util.GetStaticFile(path + name)
	reader := strings.NewReader(script)
	if err := ctrl.ssh.CopyTo(reader, "/tmp/"+path, name, "0775", len(script)); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) Install() (err error) {
	// Connect to server
	verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// Copy installation scripts to remote host
	verbose("Copying install files to server")
	if err = ctrl.CopyScript("", "controller_install_node.sh"); err != nil {
		return err
	}
	if err = ctrl.CopyScript("", "controller_install_iofog.sh"); err != nil {
		return err
	}

	// Copy service scripts to remote host
	verbose("Copying service files to server")
	if _, err = ctrl.ssh.Run("mkdir -p /tmp/iofog-controller-service"); err != nil {
		return err
	}
	if err = ctrl.CopyScript("iofog-controller-service/", "iofog-controller.initctl"); err != nil {
		return err
	}
	if err = ctrl.CopyScript("iofog-controller-service/", "iofog-controller.systemd"); err != nil {
		return err
	}
	if err = ctrl.CopyScript("iofog-controller-service/", "iofog-controller.update-rc"); err != nil {
		return err
	}

	// Define commands
	dbArgs := ""
	if ctrl.db.host != "" {
		db := ctrl.db
		dbArgs = fmt.Sprintf(" %s %s %s %s %d %s", db.provider, db.host, db.user, db.password, db.port, db.databaseName)
	}
	cmds := []command{
		{
			cmd: "sudo /tmp/controller_install_node.sh",
			msg: "Installing Node.js on Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo /tmp/controller_install_iofog.sh %s %s %s", ctrl.Version, ctrl.Repo, ctrl.Token) + dbArgs,
			msg: "Installing ioFog on Controller " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		verbose(cmd.msg)
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
	verbose("Waiting for Controller " + ctrl.Host)
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
	ctrlClient := client.New(endpoint)

	connected := false
	seconds := 0
	for !connected {
		// Time out
		if seconds > 60 {
			err = util.NewInternalError("Timed out waiting for Controller API")
			return
		}
		// Try to create the user
		if _, err = ctrlClient.GetStatus(); err != nil {
			// Retry if connection is refused, this is usually only necessary on K8s Controller
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "timeout") {
				time.Sleep(time.Millisecond * 1000)
				seconds = seconds + 1
				continue
			}
			// Return the error otherwise
			return
		}
		// No error, connected
		connected = true
		continue
	}

	return
}
