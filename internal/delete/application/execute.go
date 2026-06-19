package deleteapplication

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

func Execute(namespace, name string) error {
	// Get executor
	exe, _ := NewExecutor(namespace, name)

	// Execute deletion
	if err := exe.Execute(); err != nil {
		return err
	}

	// Leave this here as a note on general practice with Execute functions
	return config.Flush()
}
