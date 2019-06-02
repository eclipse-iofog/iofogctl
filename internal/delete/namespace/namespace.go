package deletemicroservice

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
)

func Execute(name string) error {
	// Update configuration
	err := config.DeleteNamespace(name)
	if err != nil {
		return err
	}

	// TODO (Serge) Handle config file error, retry..?

	fmt.Printf("\nNamespace %s successfully deleted.\n", name)

	return nil
}
