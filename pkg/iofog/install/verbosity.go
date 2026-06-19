package install

import (
	"fmt"
)

// Toggle HTTP output
var isVerbose bool

func IsVerbose() bool {
	return isVerbose
}

func SetVerbosity(verbose bool) {
	isVerbose = verbose
}

func Verbose(msg string) {
	if isVerbose {
		fmt.Println(msg)
	}
}
