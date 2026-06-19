package util

import (
	"fmt"
	"os"
	"sync"
)

// These are the colors we'll use in pretty printing output
const NoFormat = "\033[0m"
const CSkyblue = "\033[38;5;117m"
const CDeepskyblue = "\033[48;5;25m"
const Red = "\033[38;5;1m"
const Green = "\033[38;5;28m"

var progressPrintMu sync.Mutex

// Print a 'message' with CSkyblue color text
func PrintInfo(message string) {
	wasRunning := SpinPause()
	message = FirstToUpper(message)
	fmt.Printf(CSkyblue + message + NoFormat + "\n")
	if wasRunning {
		SpinUnpause()
	}
}

// Print 'message' with CDeepskyblue color text and background
func PrintNotify(message string) {
	wasRunning := SpinPause()
	message = FirstToUpper(message)
	fmt.Fprintf(os.Stderr, CSkyblue+"! "+message+NoFormat+"\n")
	if wasRunning {
		SpinUnpause()
	}
}

// PrintProgress prints inline progress updates with color and handles spinner state.
func PrintProgress(label string, percent int, done bool) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	progressPrintMu.Lock()
	defer progressPrintMu.Unlock()

	fmt.Printf("\r%s%s: %3d%%%s", CSkyblue, label, percent, NoFormat)
	if done {
		fmt.Print("\n")
	}
}

// Print 'message' with green color text
func PrintSuccess(message string) {
	SpinStop()
	message = FirstToUpper(message)
	fmt.Printf(Green + "✔ " + message + NoFormat + "\n")
}

// Print 'message' with red color text
func PrintError(message string) {
	SpinStop()
	message = FirstToUpper(message)
	fmt.Fprintf(os.Stderr, Red+"✘ "+message+NoFormat+"\n")
}
