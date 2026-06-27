package cmd

import (
	"fmt"
)

var pkg struct {
	flagDescDetached string
	flagDescYaml     string
	succMove         string
}

func init() {
	pkg.flagDescDetached = "Specify command is to run against detached resources"
	pkg.flagDescYaml = "YAML file containing specifications for ioFog resources to deploy"
	pkg.succMove = "Successfully moved %s %s to %s %s"
}

func getMoveSuccessMessage(resource, name, otherResource, otherName string) string {
	return fmt.Sprintf(pkg.succMove, resource, name, otherResource, otherName)
}
