package createnamespace

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name string) error {
	if err := util.IsLowerAlphanumeric("Namespace", name); err != nil {
		return err
	}
	// Update configuration
	err := config.AddNamespace(name, util.NowUTC())
	if err != nil {
		return err
	}

	return config.Flush()
}
