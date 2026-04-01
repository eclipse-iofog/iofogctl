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
	isInPrompt     bool // New variable to track if we're in a prompt state
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
	// spin.Stop()
	spin.Suffix = " " + msg + "\n"
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

// SpinHandlePrompt pauses the spinner when a prompt is about to be shown
func SpinHandlePrompt() {
	if isRunning {
		SpinPause()
		isInPrompt = true
	}
}

// SpinHandlePromptComplete resumes the spinner after a prompt is handled
func SpinHandlePromptComplete() {
	if isInPrompt {
		SpinUnpause()
		isInPrompt = false
	}
}
