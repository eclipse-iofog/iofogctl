package get

import (
	"os"
	"fmt"
	"text/tabwriter"
)

type row struct {
	name string
	status string
	age string
}

func print(rows []row) error {
	minWidth := 16
	tabWidth := 8
	padding := 0
	writer := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, '\t', 0)
	defer writer.Flush()

	headers := [3]string{"NAME", "STATUS", "AGE"}
	_, err := fmt.Fprintf(writer, "\n%s\t%s\t%s\t\n", headers[0], headers[1], headers[2])
	if err != nil {
		return err
	}

	for _, row := range rows {
		_, err = fmt.Fprintf(writer, "%s\t%s\t%s\t\n", row.name, row.status, row.age)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(writer, "\n")
	return err
}