package controller

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string) error {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	// Check that Controller exists in current namespace
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Get the Controller to rename
	controller, err := controlPlane.GetController(name)
	if err != nil {
		return err
	}

	// Check new name is valid
	if err := util.IsLowerAlphanumeric("Controller", newName); err != nil {
		return err
	}

	// Perform the rename
	util.SpinStart(fmt.Sprintf("Renaming Controller %s", name))
	controller.SetName(newName)
	ns.SetControlPlane(controlPlane)

	return config.Flush()
}
