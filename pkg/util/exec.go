package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Exec command
func Exec(env, cmdName string, args ...string) (stdout bytes.Buffer, err error) {
	if IsDebug() {
		fmt.Printf("[LOCAL]: Running: %s %s\n", cmdName, strings.Join(args, " "))
	}

	// Instantiate command object
	cmd := exec.Command(cmdName, args...)

	// Instantiate output objects
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set env vars
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env)

	// Run command
	err = cmd.Run()
	if err != nil {
		if IsDebug() && stderr.Len() > 0 {
			fmt.Printf("[LOCAL]: stderr: %s\n", strings.TrimSpace(stderr.String()))
		}
		err = NewInternalError(stderr.String())
		return
	}
	if IsDebug() && stdout.Len() > 0 {
		fmt.Printf("[LOCAL]: stdout: %s\n", strings.TrimSpace(stdout.String()))
	}
	return
}
