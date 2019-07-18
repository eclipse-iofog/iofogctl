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

package util

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type SecureShellClient struct {
	user            string
	host            string
	port            int
	privKeyFilename string
	config          *ssh.ClientConfig
	conn            *ssh.Client
}

func NewSecureShellClient(user, host, privKeyFilename string) *SecureShellClient {
	return &SecureShellClient{
		user:            user,
		host:            host,
		port:            22,
		privKeyFilename: privKeyFilename,
	}
}

func (cl *SecureShellClient) SetPort(port int) {
	cl.port = port
}

func (cl *SecureShellClient) Connect() (err error) {
	// Don't bother connecting twice
	if cl.conn != nil {
		return nil
	}

	// Parse keys
	key, err := cl.getPublicKey()
	if err != nil {
		return err
	}

	// Instantiate config
	cl.config = &ssh.ClientConfig{
		User: cl.user,
		Auth: []ssh.AuthMethod{
			key,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect
	endpoint := cl.host + ":" + strconv.Itoa(cl.port)
	cl.conn, err = ssh.Dial("tcp", endpoint, cl.config)
	if err != nil {
		return err
	}

	return nil
}

func (cl *SecureShellClient) Disconnect() error {
	if cl.conn == nil {
		return nil
	}

	err := cl.conn.Close()
	if err != nil {
		return err
	}
	cl.conn = nil
	return nil
}

func (cl *SecureShellClient) Run(cmd string) (stdout bytes.Buffer, err error) {
	// Establish the session
	session, err := cl.conn.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	// Connect pipes
	session.Stdout = &stdout
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}

	// Run the command
	err = session.Run(cmd)
	if err != nil {
		logFile := "/tmp/iofog.log"
		errorSuffix := "stdout has been appended to " + logFile
		if err = ioutil.WriteFile(logFile, stdout.Bytes(), 0644); err != nil {
			errorSuffix = "Failed to append stdout to log file"
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(stderr)
		err = NewInternalError("Error during SSH session\nstderr: " + buf.String() + errorSuffix)
		return
	}
	return
}

func (cl *SecureShellClient) getPublicKey() (authMeth ssh.AuthMethod, err error) {
	// Read priv key file, MUST BE RSA
	key, err := ioutil.ReadFile(cl.privKeyFilename)
	if err != nil {
		return
	}

	// Parse key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}

	// Return pubkey obj
	authMeth = ssh.PublicKeys(signer)

	return
}

func (cl *SecureShellClient) RunUntil(condition *regexp.Regexp, cmd string) (err error) {
	// Establish the session
	session, err := cl.conn.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	// Connect pipes
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}
	for iter := 0; iter < 30; iter++ {
		// Refresh stdout for every iter
		stdoutBuffer := bytes.Buffer{}
		session.Stdout = &stdoutBuffer

		// Run the command
		err = session.Run(cmd)
		if err != nil {
			logFile := "/tmp/iofog.log"
			errorSuffix := "stdout has been appended to " + logFile
			if err = ioutil.WriteFile(logFile, stdoutBuffer.Bytes(), 0644); err != nil {
				errorSuffix = "Failed to append stdout to log file"
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(stderr)
			err = NewInternalError("Error during SSH session\nstderr: " + buf.String() + errorSuffix)
			return
		}
		if condition.MatchString(stdoutBuffer.String()) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return NewInternalError("Timed out waiting for condition '" + condition.String() + "' with SSH command: " + cmd)
}
