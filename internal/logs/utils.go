package logs

import (
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func printContainerLogs(stdout, stderr string) {
	util.WriteStdoutString(stdout)
	util.WriteStderrString(stderr)
}
