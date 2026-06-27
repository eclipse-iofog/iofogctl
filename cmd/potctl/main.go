package main

import (
	"github.com/eclipse-iofog/iofogctl/internal/cmd"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func main() {
	config.Init("")
	rootCmd := cmd.NewRootCommand()
	err := rootCmd.Execute()
	util.Check(err)
}
