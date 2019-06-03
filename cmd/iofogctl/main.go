package main

import (
	"github.com/eclipse-iofog/cli/internal/cmd"
	"github.com/eclipse-iofog/cli/pkg/util"
)

func main() {
	root := cmd.NewRootCommand()
	err := root.Execute()
	util.Check(err)
}
