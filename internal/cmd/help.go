package cmd

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// ex formats Cobra help strings with the build-time CLI binary name as %[1]s.
func ex(format string, args ...any) string {
	all := append([]any{util.GetCliBinaryName()}, args...)
	return fmt.Sprintf(format, all...)
}
