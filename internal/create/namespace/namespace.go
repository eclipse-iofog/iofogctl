package createnamespace

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

func Execute(name string) error {
	// Update configuration
	err := config.AddNamespace(name)
	if err != nil {
		return err
	}

	fmt.Printf("\nNamespace %s successfully created.\n", name)

	return nil
}
