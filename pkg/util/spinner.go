package util

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

var (
	quiet          bool
	spin           *spinner.Spinner // There is only one spinner, output overlaps with multiple concurrent spinners
	currentMessage string
	isRunning      bool
)

func init() {
	// Note: don't set the colour here, it will display the spinner when you don't want it to
	spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
}

func SpinEnable(isEnabled bool) {
	quiet = !isEnabled
}

func SpinStart(msg string) {
	isRunning = true
	currentMessage = msg
	if quiet {
		fmt.Println(msg)
		return
	}
	_ = spin.Color("red")
	spin.Stop()
	spin.Suffix = " " + msg
	spin.Start()
}

func SpinPause() bool {
	wasRunning := isRunning
	SpinStop()
	return wasRunning
}

func SpinUnpause() {
	SpinStart(currentMessage)
}

func SpinStop() {
	isRunning = false
	if quiet {
		return
	}
	spin.Stop()
}
