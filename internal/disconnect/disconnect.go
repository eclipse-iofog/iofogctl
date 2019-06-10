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

	// Wipe the namespace
	err = config.DeleteNamespace(opt.Namespace)
	if err != nil {
		return err
	}
	err = config.AddNamespace(opt.Namespace, ns.Created)
	if err != nil {
		return err
	}

	return config.Flush()
}
