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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	maxFileSize     = 100 * 1024 * 1024 // 100MB
	scpProgressStep = 5 * 1024 * 1024   // 5 MB - log transfer progress every 5 MB when debug
)

type SecureShellClient struct {
	user            string
	host            string
	port            int
	privKeyFilename string
	config          *ssh.ClientConfig
	conn            *ssh.Client
}

func NewSecureShellClient(user, host, privKeyFilename string) (*SecureShellClient, error) {
	cl := &SecureShellClient{
		user:            user,
		host:            host,
		port:            22,
		privKeyFilename: privKeyFilename,
	}
	// Parse keys
	SSHVerbose("Parsing keys")
	key, err := cl.getPublicKey()
	if err != nil {
		return nil, err
	}

	// Instantiate config
	SSHVerbose("Configuring SSH client")
	cl.config = &ssh.ClientConfig{
		User: cl.user,
		Auth: []ssh.AuthMethod{
			key,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	SSHVerbose("Config:")
	SSHVerbose(fmt.Sprintf("User: %s", cl.user))
	SSHVerbose(fmt.Sprintf("Auth method: %v", key))
	return cl, nil
}

func (cl *SecureShellClient) SetPort(port int) {
	SSHVerbose(fmt.Sprintf("Setting port to %v", port))
	cl.port = port
}

func (cl *SecureShellClient) Connect() (err error) {
	// Don't bother connecting twice
	SSHVerbose("Initialiasing connection")
	if cl.conn != nil {
		return nil
	}

	// Connect
	endpoint := cl.host + ":" + strconv.Itoa(cl.port)
	SSHVerbose(fmt.Sprintf("TCP dialing %s", endpoint))
	cl.conn, err = ssh.Dial("tcp", endpoint, cl.config)
	if err != nil {
		return err
	}

	return nil
}

func (cl *SecureShellClient) Disconnect() error {
	SSHVerbose("Disconnecting...")
	if cl.conn == nil {
		return nil
	}

	err := cl.conn.Close()
	if err != nil {
		return err
	}

	SSHVerbose("Connection closed")
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
		err = format(err, nil, readToBuffer(stderr))
		return
	}

	// Run the command
	SSHVerbose(fmt.Sprintf("Running: %s", cmd))
	err = session.Run(cmd)
	if err != nil {
		err = format(err, &stdout, readToBuffer(stderr))
		return
	}
	return
}

func format(err error, stdout, stderr fmt.Stringer) error {
	if err == nil {
		return nil
	}
	msg := "Error during SSH Session"
	if stdout != nil && stdout.String() != "" {
		msg = fmt.Sprintf("%s\n%s", msg, stdout.String())
	}
	if stderr != nil && stderr.String() != "" {
		msg = fmt.Sprintf("%s\n%s", msg, stderr.String())
	}
	msg = fmt.Sprintf("%s\n%s", msg, err.Error())

	return errors.New(msg)
}

func (cl *SecureShellClient) getPublicKey() (authMeth ssh.AuthMethod, err error) {
	// Read priv key file, MUST BE RSA
	SSHVerbose(fmt.Sprintf("Reading private key: %s", cl.privKeyFilename))
	key, err := os.ReadFile(cl.privKeyFilename)
	if err != nil {
		return
	}

	// Parse key
	SSHVerbose("Parsing key")
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}

	// Return pubkey obj
	SSHVerbose("Creating auth method based on key pair")
	authMeth = ssh.PublicKeys(signer)

	return
}

func (cl *SecureShellClient) RunUntil(condition *regexp.Regexp, cmd string, ignoredErrors []string) (err error) {
	// Retry until string condition matches
	for iter := 0; iter < 30; iter++ {
		SSHVerbose(fmt.Sprintf("Try %v", iter))
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
		stdoutBuffer := &bytes.Buffer{}
		session.Stdout = stdoutBuffer

		// Run the command
		SSHVerbose(fmt.Sprintf("Running: %s", cmd))
		err = session.Run(cmd)
		// Ignore specified errors
		if err != nil {
			errMsg := err.Error()
			for _, toIgnore := range ignoredErrors {
				if strings.Contains(errMsg, toIgnore) {
					// ignore error
					SSHVerbose(fmt.Sprintf("Ignored error: %s", errMsg))
					err = nil
					break
				}
			}
		}
		if err != nil {
			return format(err, stdoutBuffer, readToBuffer(stderr))
		}
		if condition.MatchString(stdoutBuffer.String()) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return NewInternalError("Timed out waiting for condition '" + condition.String() + "' with SSH command: " + cmd)
}

// progressReader wraps a reader and logs SFTP transfer progress every 5 MB when debug is on.
type progressReader struct {
	r          io.Reader
	total      int64
	n          int64
	lastLogged int64
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.n += int64(n)
		for p.n-p.lastLogged >= scpProgressStep {
			p.lastLogged += scpProgressStep
			mbSent := p.lastLogged / (1024 * 1024)
			mbTotal := p.total / (1024 * 1024)
			pct := int64(0)
			if p.total > 0 {
				pct = (p.lastLogged * 100) / p.total
			}
			SSHVerbose(fmt.Sprintf("SFTP: sent %d MB / %d MB (%d%%)", mbSent, mbTotal, pct))
		}
		if p.n >= p.total && p.total > 0 && p.lastLogged < p.total {
			mbTotal := p.total / (1024 * 1024)
			SSHVerbose(fmt.Sprintf("SFTP: sent %d MB / %d MB (100%%)", mbTotal, mbTotal))
			p.lastLogged = p.total
		}
	}
	return n, err
}

func (cl *SecureShellClient) CopyTo(reader io.Reader, destPath, destFilename, permissions string, size int64) error {
	SSHVerbose(fmt.Sprintf("Copying file %s...", JoinAgentPath(destPath, destFilename)))
	if !regexp.MustCompile(`\d{4}`).MatchString(permissions) {
		return NewError("Invalid file permission specified: " + permissions)
	}
	if cl.conn == nil {
		return NewError("SSH connection not established; call Connect() before CopyTo")
	}

	sftpClient, err := sftp.NewClient(cl.conn)
	if err != nil {
		return fmt.Errorf("SFTP subsystem not available on remote host: %w", err)
	}
	defer sftpClient.Close()

	remotePath := JoinAgentPath(destPath, destFilename)
	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("SFTP create %s: %w", remotePath, err)
	}

	copyReader := io.Reader(reader)
	if IsDebug() && size > 0 {
		copyReader = &progressReader{r: reader, total: size}
	}
	if _, err := io.Copy(dstFile, copyReader); err != nil {
		_ = dstFile.Close()
		return fmt.Errorf("SFTP write %s: %w", remotePath, err)
	}
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("SFTP close %s: %w", remotePath, err)
	}

	perm, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions %s: %w", permissions, err)
	}
	if err := sftpClient.Chmod(remotePath, os.FileMode(perm)); err != nil {
		return fmt.Errorf("SFTP chmod %s: %w", remotePath, err)
	}
	return nil
}

func (cl *SecureShellClient) CopyFolderTo(srcPath, destPath, permissions string, recurse bool) error {
	SSHVerbose("Copying folder...")
	files, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() && recurse {
			// Create the dir if necessary
			if err := cl.CreateFolder(JoinAgentPath(destPath, file.Name())); err != nil {
				return err
			}
			// Copy contents of dir
			if err := cl.CopyFolderTo(
				filepath.Join(srcPath, file.Name()),
				JoinAgentPath(destPath, file.Name()),
				permissions,
				true,
			); err != nil {
				return err
			}
		} else {
			// Read the file
			openFile, err := os.Open(filepath.Join(srcPath, file.Name()))
			if err != nil {
				return err
			}
			fileInfo, err := openFile.Stat()
			if err != nil {
				return err
			}
			if fileInfo.Size() > maxFileSize {
				return fmt.Errorf("file %s is too large (max size: %d bytes)", fileInfo.Name(), maxFileSize)
			}
			// Copy the file
			if err := cl.CopyTo(openFile, destPath, file.Name(), addLeadingZero(permissions), fileInfo.Size()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cl *SecureShellClient) CreateFolder(path string) error {
	path = AddTrailingSlash(path)
	SSHVerbose(fmt.Sprintf("Creating folder %s", path))
	SSHVerbose(fmt.Sprintf("Running: %s", "mkdir -p "+path))
	if _, err := cl.Run("mkdir -p " + path); err != nil {
		if strings.Contains(err.Error(), "exists") {
			return nil
		}
		// Retry with sudo
		if strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			if _, sudoErr := cl.Run("sudo -S mkdir -p " + path); sudoErr != nil {
				if !strings.Contains(sudoErr.Error(), "exists") {
					return sudoErr
				}
			}
		}
		return err
	}
	return nil
}

func addLeadingZero(in string) string {
	if in[0:0] != "0" {
		in = "0" + in
	}
	return in
}

func AddTrailingSlash(in string) string {
	if in[len(in)-1:] != "/" {
		in += "/"
	}
	return in
}

func SSHVerbose(msg string) {
	if IsDebug() {
		fmt.Printf("[SSH]: %s\n", msg)
	}
}

func JoinAgentPath(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func readToBuffer(reader io.Reader) (buf *bytes.Buffer) {
	buf = new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		buf = nil
		return
	}
	return
}
