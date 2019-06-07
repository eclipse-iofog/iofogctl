package logs

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
}

func newControllerExecutor(namespace, name string) *controllerExecutor {
	exe := &controllerExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *controllerExecutor) Execute() error {
	// Get controller config
	ctrl, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Local
	if ctrl.Host == "localhost" {
		return util.NewInternalError("Not Implemented")
	}

	// K8s
	if ctrl.KubeConfig != "" {
		out, err := util.Exec("KUBECONFIG="+ctrl.KubeConfig, "kubectl", "logs", "-l", "name=controller", "-n", "iofog")
		if err != nil {
			return err
		}
		println(out.String())
		return nil
	}

	// Remote
	ssh := util.NewSecureShellClient(ctrl.User, ctrl.Host, ctrl.KeyFile)
	err = ssh.Connect()
	if err != nil {
		return err
	}

	// Get logs
	out, err := ssh.Run("sudo cat /var/log/iofog-controller/*")
	if err != nil {
		return err
	}
	println(out.String())

	return nil
}
