package deletemicroservice

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	delete "github.com/eclipse-iofog/iofogctl/internal/delete/all"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name string, force bool) error {
	// Disallow deletion of default
	if name == "default" {
		return util.NewInputError("Cannot delete namespace named \"default\"")
	}

	// Get config
	ns, err := config.GetNamespace(name)
	if err != nil {
		return err
	}

	// Check resources exist
	hasAgents := len(ns.GetAgents()) > 0
	hasControllers := len(ns.GetControllers()) > 0

	// Force must be specified
	if !force && (hasAgents || hasControllers) {
		return util.NewInputError("Namespace " + name + " not empty. You must force the deletion if the namespace is not empty")
	}

	// Handle delete all
	if force && (hasAgents || hasControllers) {
		if err := delete.Execute(name, false, force); err != nil {
			return err
		}
	}

	// Delete namespace
	if err := config.DeleteNamespace(name); err != nil {
		return err
	}

	return config.Flush()
}
