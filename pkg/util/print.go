/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package util

import (
	"fmt"
	"os"
)

// These are the colors we'll use in pretty printing output
const NoFormat = "\033[0m"
const CSkyblue = "\033[38;5;117m"
const CDeepskyblue = "\033[48;5;25m"
const Red = "\033[38;5;1m"
const Green = "\033[38;5;28m"

// Print a 'message' with CSkyblue color text
func PrintInfo(message string) {
	msg := SpinPause()
	fmt.Printf(CSkyblue + message + NoFormat + "\n")
	SpinStart(msg)
}

// Print 'message' with CDeepskyblue color text and background
func PrintNotify(message string) {
	msg := SpinPause()
	fmt.Fprintf(os.Stderr, CSkyblue+"! "+message+NoFormat+"\n")
	SpinStart(msg)
}

// Print 'message' with green color text
func PrintSuccess(message string) {
	SpinStop()
	fmt.Printf(Green + "✔ " + message + NoFormat + "\n")
}

// Print 'message' with red color text
func PrintError(message string) {
	SpinStop()
	fmt.Fprintf(os.Stderr, Red+"✘ "+message+NoFormat+"\n")
}
