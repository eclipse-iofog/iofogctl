package deletecontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
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

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)
	}
	return err
}
