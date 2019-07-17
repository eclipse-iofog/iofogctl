package util

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

var (
	Quiet bool
	spin  *spinner.Spinner // There is only one spinner, output overlaps with multiple concurrent spinners
)

func init() {
	// Note: don't set the colour here, it will display the spinner when you don't want it to
	spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
}

func SpinStart(msg string) {
	if Quiet {
		fmt.Println(msg)
		return
	}
	spin.Color("red")
	spin.Stop()
	spin.Suffix = " " + msg
	spin.Start()
}

func SpinStop() {
	if Quiet {
		return
	}
	spin.Stop()
}
