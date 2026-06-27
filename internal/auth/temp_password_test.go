package auth

import (
	"testing"
	"unicode"

	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
	"github.com/stretchr/testify/require"
)

func TestGenerateTempPasswordMeetsComplexity(t *testing.T) {
	for i := 0; i < 20; i++ {
		password, err := generateTempPassword()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(password), 12)
		require.NoError(t, inputvalidate.ValidatePasswordComplexity(password))

		hasUpper := false
		hasDigit := false
		for _, r := range password {
			if unicode.IsUpper(r) {
				hasUpper = true
			}
			if unicode.IsDigit(r) {
				hasDigit = true
			}
		}
		require.True(t, hasUpper)
		require.True(t, hasDigit)
	}
}
