package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePasswordComplexity(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, ValidatePasswordComplexity("LocalTest12!"))
	})

	t.Run("too short", func(t *testing.T) {
		require.Error(t, ValidatePasswordComplexity("Short1!"))
	})

	t.Run("no uppercase", func(t *testing.T) {
		require.Error(t, ValidatePasswordComplexity("localtest12!"))
	})

	t.Run("no special", func(t *testing.T) {
		require.Error(t, ValidatePasswordComplexity("LocalTest1234"))
	})
}
