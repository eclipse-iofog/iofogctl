package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name string) error {
	util.SpinStart("Deleting Controller ")

	// Get executor
	exe, err := NewExecutor(namespace, name)
	if err != nil {
		return err
	}

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	return config.Flush()
}
