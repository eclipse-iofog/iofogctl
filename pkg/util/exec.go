package util

import (
	"bytes"
	"os"
	"os/exec"
)

// Exec command
func Exec(env, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env)
	err := cmd.Run()
	if err != nil {
		return NewInternalError(stderr.String())
	}
	//println(out.String())
	return nil
}
