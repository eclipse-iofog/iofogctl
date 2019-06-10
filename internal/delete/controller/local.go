package deletecontroller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type localExecutor struct {
	namespace string
	name      string
}

func newLocalExecutor(namespace, name string) *localExecutor {
	exe := &localExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *localExecutor) Execute() error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	err := config.DeleteController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
