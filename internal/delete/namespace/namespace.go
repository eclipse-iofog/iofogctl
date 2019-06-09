package deletemicroservice

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name string) error {
	// Get config
	ns, err := config.GetNamespace(name)
	if err != nil {
		return err
	}

	// Check resources exist
	hasAgents := len(ns.Agents) > 0
	hasControllers := len(ns.Controllers) > 0
	hasMicroservices := len(ns.Microservices) > 0
	if hasAgents || hasControllers || hasMicroservices {
		return util.NewInputError("Namespace " + name + " not empty")
	}

	// Delete namespace
	err = config.DeleteNamespace(name)
	if err != nil {
		return err
	}

	fmt.Printf("\nNamespace %s successfully deleted.\n", name)

	return nil
}
