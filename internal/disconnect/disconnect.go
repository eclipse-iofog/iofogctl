package disconnect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
}

func Execute(opt *Options) error {
	// Check Namespace exists
	if _, err := config.GetNamespace(opt.Namespace); err != nil {
		if util.IsNotFoundError(err) {
			// Not found, disconnection is idempotent
			return nil
		}
		// Error was not 'not found'
		return err
	}

	if err := config.DeleteNamespace(opt.Namespace); err != nil {
		return err
	}
	return config.Flush()
}
