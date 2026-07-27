package deployremotecontrolplane

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func preflightControlPlaneSSHHosts(cp *rsc.RemoteControlPlane) error {
	if cp == nil || len(cp.Controllers) == 0 {
		return nil
	}

	targets := make([]util.SSHTarget, 0, len(cp.Controllers))
	for idx := range cp.Controllers {
		ctrl := &cp.Controllers[idx]
		if err := ctrl.ValidateSSH(); err != nil {
			return err
		}
		port := ctrl.SSH.Port
		if port == 0 {
			port = 22
		}
		targets = append(targets, util.SSHTarget{
			User:    ctrl.SSH.User,
			Host:    ctrl.Host,
			Port:    port,
			KeyFile: ctrl.SSH.KeyFile,
			Label:   ctrl.Name,
		})
	}

	if err := util.PreflightSSHHosts(targets); err != nil {
		return err
	}
	util.PrintInfo("SSH preflight complete for controller hosts")
	return nil
}
