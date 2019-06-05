package get

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func print(table [][]string) error {
	minWidth := 16
	tabWidth := 8
	padding := 0
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
