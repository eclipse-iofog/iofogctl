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

import "fmt"

//
// These are the colors we'll use in printing output
//
const NoFormat = "\033[0m"
const CSkyblue = "\033[38;5;117m"
const CDeepskyblue = "\033[48;5;25m"
const Red = "\033[38;5;1m"
const Green = "\033[38;5;28m"

//
// Print a 'message' with CSkyblue color text
//
func PrintInfo(message string) {
	fmt.Printf(CSkyblue + message + NoFormat + "\n")
}

//
// Print 'message' with CDeepskyblue color text and background
//
func PrintNotify(message string) {
	fmt.Printf(CDeepskyblue + message + NoFormat + "\n")
}

//
// Print 'message' with green color text
//
func PrintSucess(message string) {
	fmt.Printf(Green + message + NoFormat + "\n")
}

//
// Print 'message' with red color text
//
func PrintError(message string) {
	fmt.Printf(Red + message + NoFormat + "\n")
}
