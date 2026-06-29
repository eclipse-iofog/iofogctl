package authgroup

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// Table builds a tabular representation of auth groups.
func Table(groups []client.AuthGroupResponse) [][]string {
	table := make([][]string, len(groups)+1)
	table[0] = []string{"NAME", "SYSTEM", "MFA"}
	for idx, group := range groups {
		table[idx+1] = []string{
			group.Name,
			formatBool(group.IsSystem),
			formatBool(group.MfaRequired),
		}
	}
	return table
}
