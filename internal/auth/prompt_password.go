package auth

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"golang.org/x/term"
)

var (
	readPasswordFn           = readPasswordHidden
	isTerminalFn             = term.IsTerminal
	stdin                    = os.Stdin
	stdout         io.Writer = os.Stdout
)

// PromptIofogUserPassword interactively collects and confirms an iofogUser password.
func PromptIofogUserPassword(email string) (string, error) {
	if !isTerminalFn(int(stdin.Fd())) {
		return "", util.NewInputError("iofogUser.password is required in non-interactive mode")
	}

	util.SpinHandlePrompt()
	defer util.SpinHandlePromptComplete()

	for {
		password, err := readPasswordFn(fmt.Sprintf("Enter password for iofog user %s: ", email))
		if err != nil {
			return "", err
		}
		confirm, err := readPasswordFn("Confirm password: ")
		if err != nil {
			return "", err
		}
		if password != confirm {
			fmt.Fprintln(stdout, "Passwords do not match.")
			continue
		}
		if err := inputvalidate.ValidatePasswordComplexity(password); err != nil {
			fmt.Fprintln(stdout, err.Error())
			continue
		}
		return password, nil
	}
}

func readPasswordHidden(prompt string) (string, error) {
	fmt.Fprint(stdout, prompt)
	defer fmt.Fprintln(stdout)

	reader := bufio.NewReader(stdin)
	if isTerminalFn(int(stdin.Fd())) {
		bytes, err := term.ReadPassword(int(stdin.Fd()))
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
