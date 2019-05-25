package get

import (
	"github.com/eclipse-iofog/cli/pkg/config"
	"github.com/eclipse-iofog/cli/pkg/util"
)
type get struct {
	configManager *config.Manager
}

func new() *get {
	g := &get{}
	g.configManager = config.NewManager()
	return g
}

func (get *get) execute(resource string) error {
	println("Execute get")
	switch resource {
	case "controllers":
		//
	case "agents":
		//
	case "microservices":
		//
	default:
		// Print out error
		return util.NewNotFound("")
	}
	return nil
}