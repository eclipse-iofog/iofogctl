package disconnect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type Options struct {
	Namespace string
	Name      string
}

func Execute(opt *Options) error {
	// Check namespace
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return err
	}

	// TODO: Implement this if we can share resources between users
	// Connect to controller
	//if len(ns.Controllers) != 1 {
	//	return util.NewInternalError("Expected one controller in namespace " + opt.Namespace)
	//}
	// Delete user

	// Wipe the namespace
	err = config.DeleteNamespace(opt.Namespace)
	if err != nil {
		return err
	}
	err = config.AddNamespace(opt.Namespace, ns.Created)
	if err != nil {
		return err
	}

	return nil
}
