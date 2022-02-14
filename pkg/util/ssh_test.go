package util

import (
	"fmt"
	"testing"
)

func TestSSH(t *testing.T) {
	ssh, err := NewSecureShellClient("neha", "34.145.96.246", "/Users/nehanaithani/.ssh/id_rsa")
	if err != nil {
		fmt.Errorf(err.Error())
	}
	ssh.SetPort(22)
	if err := ssh.Connect(); err != nil {
		fmt.Errorf(err.Error())
	}
	defer ssh.Disconnect()
	out, err := ssh.Run("sudo /etc/iofog/agent/install_iofog.sh 3.0.0_dev_b17805 iofog/iofog-agent-dev")
	if err != nil {
		fmt.Errorf(err.Error())
	}
	fmt.Printf(out.String())
}
