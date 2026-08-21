package get

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func print(table [][]string) error {
	// Space padding so column alignment does not depend on the terminal tab width.
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
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
