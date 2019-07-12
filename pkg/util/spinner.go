package util

import (
	"github.com/briandowns/spinner"
	"time"
)

// There is only one spinner, output overlaps with multiple concurrent spinners
var spin *spinner.Spinner

func init() {
	// Note: don't set the colour here, it will display the spinner when you don't want it to
	spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
}

func SpinStart(msg string) {
	spin.Color("red")
	spin.Stop()
	spin.Suffix = " " + msg
	spin.Start()
}

func SpinStop() {
	spin.Stop()
}
