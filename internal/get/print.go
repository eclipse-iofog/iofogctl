/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package get

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func print(table [][]string) error {
	minWidth := 16
	tabWidth := 8
	padding := 1
	writer := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, '\t', 0)
	defer writer.Flush()

	for _, row := range table {
		for _, col := range row {
			_, err := fmt.Fprintf(writer, "%s\t", col)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(writer, "\n")
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(writer, "\n")
	if err != nil {
		return err
	}

	return nil
}

func printNamespace(namespace string) {
	fmt.Printf("NAMESPACE\n%s\n\n", namespace)
}
