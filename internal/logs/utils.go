package logs

import (
	"os"
)

func printContainerLogs(stdout, stderr string) {
	os.Stdout.WriteString(stdout)
	os.Stderr.WriteString(stderr)
}
