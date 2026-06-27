package install

import (
	"fmt"
	"strings"
)

// AgentProcedures holds layered install script entrypoints shared by edgelet and controller flows.
type AgentProcedures struct {
	check          Entrypoint `yaml:"-"`
	Deps           Entrypoint `yaml:"deps,omitempty"`
	Install        Entrypoint `yaml:"install,omitempty"`
	Uninstall      Entrypoint `yaml:"uninstall,omitempty"`
	scriptNames    []string   `yaml:"-"`
	scriptContents []string   `yaml:"-"`
}

// Entrypoint describes one embedded or custom install script invocation.
type Entrypoint struct {
	Name     string   `yaml:"entrypoint"`
	Args     []string `yaml:"args"`
	destPath string   `yaml:"-"`
}

func (script *Entrypoint) getCommand() string {
	if script.destPath == "" {
		return ""
	}
	if len(script.Args) == 0 {
		return script.destPath
	}
	return fmt.Sprintf("%s %s", script.destPath, shellJoinArgs(script.Args))
}

func shellJoinArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuoteArg(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuoteArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
}

type command struct {
	cmd string
	msg string
}
