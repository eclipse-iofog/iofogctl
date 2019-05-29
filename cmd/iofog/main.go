package main

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

func main() {
	cmd := newRootCommand()
	err := cmd.Execute()
	util.Check(err)
}
