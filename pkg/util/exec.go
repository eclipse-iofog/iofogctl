package util

import (
	"bytes"
	"os"
	"os/exec"
)

// Exec command
func Exec(env, name string, args ...string) (bytes.Buffer, error) {
	// Instantiate command object
	cmd := exec.Command(name, args...)

	// Instantiate output objects
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set env vars
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env)

	// Run command
	err := cmd.Run()
	if err != nil {
		return stdout, NewInternalError(stderr.String())
	}
	return stdout, nil
}
