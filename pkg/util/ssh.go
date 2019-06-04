package util

import (
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
)

type SecureShellClient struct {
	user            string
	host            string
	privKeyFilename string
	config          *ssh.ClientConfig
	conn            *ssh.Client
}

func NewSecureShellClient(user, host, privKeyFilename string) *SecureShellClient {
	return &SecureShellClient{
		user:            user,
		host:            host,
		privKeyFilename: privKeyFilename,
	}
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
	cl.conn, err = ssh.Dial("tcp", cl.host+":22", cl.config)
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

func (cl *SecureShellClient) Run(cmd string) error {
	session, err := cl.conn.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	sessStdOut, err := session.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stdout, sessStdOut)
	sessStderr, err := session.StderrPipe()
	if err != nil {
		panic(err)
	}
	go io.Copy(os.Stderr, sessStderr)

	err = session.Run(cmd)
	if err != nil {
		return err
	}
	return nil
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
