package namespace

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(name, newName string) error {
	if name == "" || name == "default" {
		return util.NewError("Cannot rename default or nonexistant namespaces")
	}
	if err := util.IsLowerAlphanumeric("Namespace", newName); err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming Namespace %s", name))

	if err := config.RenameNamespace(name, newName); err != nil {
		return err
	}
	return config.Flush()
}
