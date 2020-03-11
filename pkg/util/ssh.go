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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
		stderrBuf := new(bytes.Buffer)
		stderrBuf.ReadFrom(stderr)
		err = format(err, &stdout, stderrBuf)
		return
	}
	return
}

func format(err error, stdout, stderr *bytes.Buffer) error {
	if err == nil || stdout == nil || stderr == nil {
		return err
	}

	msg := fmt.Sprintf("Error during SSH session\nstderr:\n%s\nstdout:\n%s", stderr.String(), stdout.String())
	return errors.New(msg)
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

func (cl *SecureShellClient) RunUntil(condition *regexp.Regexp, cmd string, ignoredErrors []string) (err error) {
	// Retry until string condition matches
	for iter := 0; iter < 30; iter++ {
		// Establish the session
		var session *ssh.Session
		session, err = cl.conn.NewSession()
		if err != nil {
			return
		}
		defer session.Close()

		// Connect pipes
		var stderr io.Reader
		stderr, err = session.StderrPipe()
		if err != nil {
			return
		}
		// Refresh stdout for every iter
		stdoutBuffer := bytes.Buffer{}
		session.Stdout = &stdoutBuffer

		// Run the command
		err = session.Run(cmd)
		// Ignore specified errors
		if err != nil {
			errMsg := err.Error()
			for _, toIgnore := range ignoredErrors {
				if strings.Contains(errMsg, toIgnore) {
					// ignore error
					err = nil
					break
				}
			}
		}
		if err != nil {
			stderrBuf := new(bytes.Buffer)
			stderrBuf.ReadFrom(stderr)
			err = format(err, &stdoutBuffer, stderrBuf)
			return
		}
		if condition.MatchString(stdoutBuffer.String()) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return NewInternalError("Timed out waiting for condition '" + condition.String() + "' with SSH command: " + cmd)
}

func (cl *SecureShellClient) CopyTo(reader io.Reader, destPath, destFilename, permissions string, size int64) error {
	// Check permissions string
	if !regexp.MustCompile(`\d{4}`).MatchString(permissions) {
		return NewError("Invalid file permission specified: " + permissions)
	}

	// Establish the session
	session, err := cl.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Start routine to write file
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Instantiate reference to stdin
		remoteStdin, err := session.StdinPipe()
		if err != nil {
			errChan <- err
		}
		defer remoteStdin.Close()

		// Write to stdin
		fmt.Fprintf(remoteStdin, "C%s %d %s\n", permissions, size, destFilename)
		io.Copy(remoteStdin, reader)
		fmt.Fprint(remoteStdin, "\x00")
	}()

	// Start the scp command
	session.Run("/usr/bin/scp -t " + destPath)

	// Wait for completion
	wg.Wait()

	// Check for errors
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (cl *SecureShellClient) CopyFolderTo(srcPath, destPath, permissions string, recurse bool) error {
	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() && recurse {
			// Create the dir if necessary
			if err := cl.CreateFolder(addTrailingSlash(destPath) + file.Name()); err != nil {
				return err
			}
			// Copy contents of dir
			return cl.CopyFolderTo(
				addTrailingSlash(srcPath)+file.Name(),
				addTrailingSlash(destPath)+file.Name(),
				permissions,
				true,
			)
		} else {
			// Read the file
			openFile, err := os.Open(addTrailingSlash(srcPath) + file.Name())
			if err != nil {
				return err
			}
			// Copy the file
			if err := cl.CopyTo(openFile, addTrailingSlash(destPath), file.Name(), addLeadingZero(permissions), file.Size()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cl *SecureShellClient) CreateFolder(path string) error {
	if _, err := cl.Run("mkdir -p " + addTrailingSlash(path)); err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return err
		}
	}
	return nil
}

func addLeadingZero(in string) string {
	if in[0:0] != "0" {
		in = "0" + in
	}
	return in
}

func addTrailingSlash(in string) string {
	if in[len(in)-1:] != "/" {
		in = in + "/"
	}
	return in
}
