package util

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

var (
	quiet          bool
	spin           *spinner.Spinner // There is only one spinner, output overlaps with multiple concurrent spinners
	currentMessage string
)

func init() {
	// Note: don't set the colour here, it will display the spinner when you don't want it to
	spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
}

func SpinEnable(isEnabled bool) {
	quiet = !isEnabled
}

func SpinStart(msg string) {
	currentMessage = msg
	if quiet {
		fmt.Println(msg)
		return
	}
	spin.Color("red")
	spin.Stop()
	spin.Suffix = " " + msg
	spin.Start()
}

func SpinPause() string {
	SpinStop()
	return currentMessage
}

func SpinStop() {
	if quiet {
		return
	}
	spin.Stop()
}
