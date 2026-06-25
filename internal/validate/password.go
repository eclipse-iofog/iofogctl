package validate

import (
	"strings"
	"unicode"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// ValidatePasswordComplexity returns an InputError when password fails v3.8 policy:
// at least 12 characters, one uppercase letter, and one special (non-alphanumeric) character.
//
//nolint:revive // ValidatePasswordComplexity matches the validate package naming convention.
func ValidatePasswordComplexity(password string) error {
	var failures []string
	if len(password) < 12 {
		failures = append(failures, "at least 12 characters")
	}
	hasUpper := false
	hasSpecial := false
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			hasSpecial = true
		}
	}
	if !hasUpper {
		failures = append(failures, "at least one uppercase letter")
	}
	if !hasSpecial {
		failures = append(failures, "at least one special character")
	}
	if len(failures) == 0 {
		return nil
	}
	return util.NewInputError("password must contain " + strings.Join(failures, ", "))
}
