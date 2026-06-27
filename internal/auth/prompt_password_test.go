package auth

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func TestPromptIofogUserPasswordNonInteractive(t *testing.T) {
	t.Cleanup(resetPromptDeps)
	isTerminalFn = func(int) bool { return false }

	_, err := PromptIofogUserPassword("user@domain.com")
	require.Error(t, err)
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}

func TestPromptIofogUserPasswordMismatchThenSuccess(t *testing.T) {
	t.Cleanup(resetPromptDeps)
	isTerminalFn = func(int) bool { return true }
	stdout = io.Discard

	responses := []string{"LocalTest12!", "WrongConfirm1!", "LocalTest12!", "LocalTest12!"}
	readPasswordFn = func(string) (string, error) {
		if len(responses) == 0 {
			return "", io.EOF
		}
		next := responses[0]
		responses = responses[1:]
		return next, nil
	}

	password, err := PromptIofogUserPassword("user@domain.com")
	require.NoError(t, err)
	require.Equal(t, "LocalTest12!", password)
}

func TestPromptIofogUserPasswordComplexityRetry(t *testing.T) {
	t.Cleanup(resetPromptDeps)
	isTerminalFn = func(int) bool { return true }
	stdout = io.Discard

	responses := []string{"short", "short", "LocalTest12!", "LocalTest12!"}
	readPasswordFn = func(string) (string, error) {
		next := responses[0]
		responses = responses[1:]
		return next, nil
	}

	password, err := PromptIofogUserPassword("user@domain.com")
	require.NoError(t, err)
	require.NoError(t, inputvalidate.ValidatePasswordComplexity(password))
}

func TestPromptIofogUserPasswordDoesNotPromptBootstrap(t *testing.T) {
	t.Cleanup(resetPromptDeps)
	var prompts bytes.Buffer
	readPasswordFn = func(prompt string) (string, error) {
		prompts.WriteString(prompt)
		return "LocalTest12!", nil
	}
	isTerminalFn = func(int) bool { return true }
	stdout = io.Discard

	_, err := PromptIofogUserPassword("user@domain.com")
	require.NoError(t, err)
	combined := prompts.String()
	require.True(t, strings.Contains(combined, "iofog user user@domain.com"))
	require.False(t, strings.Contains(strings.ToLower(combined), "bootstrap"))
}

func resetPromptDeps() {
	readPasswordFn = readPasswordHidden
	isTerminalFn = term.IsTerminal
	stdout = os.Stdout
}
