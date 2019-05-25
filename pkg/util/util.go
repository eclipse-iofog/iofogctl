package util

import (
	"os"
)

// Check export
func Check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}