package main

import (
	"github.com/eclipse-iofog/iofogctl/internal/cmd"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func main() {
	root := cmd.NewRootCommand()
	err := root.Execute()
	util.Check(err)
}
