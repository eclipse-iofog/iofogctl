package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func prepareUserAndSaveConfig(opt *Options) (configEntry config.Controller, err error) {
	var configUser config.IofogUser
	// Check for existing user
	ctrl, err := config.GetController(opt.Namespace, opt.Name)
	if opt.IofogUser.Email != "" && opt.IofogUser.Password != "" {
		// Use user provided in the yaml file
		configUser = opt.IofogUser
	} else if err == nil {
		// Use existing user
		configUser = ctrl.IofogUser
	} else {
		// Generate new user
		configUser = config.NewRandomUser()
	}

	// Record the user
	configEntry = config.Controller{
		User:       opt.User,
		Host:       opt.Host,
		Port:       opt.Port,
		KeyFile:    opt.KeyFile,
		Name:       opt.Name,
		KubeConfig: opt.KubeConfig,
		IofogUser:  configUser,
		Created:    util.NowUTC(),
	}
	if err = config.UpdateController(opt.Namespace, configEntry); err != nil {
		return
	}
	if err = config.Flush(); err != nil {
		return
	}

	return
}
