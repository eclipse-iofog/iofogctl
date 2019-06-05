package createnamespace

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/util"
)

func Execute(name string) error {
	// Update configuration
	err := config.AddNamespace(name, util.Now())
	if err != nil {
		return err
	}

	fmt.Printf("\nNamespace %s successfully created.\n", name)

	return nil
}
