package util

import (
	"os"
	"strings"
)

func ReplaceTilde(input string) (string, error) {
	if strings.Contains(input, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(input, "~", homeDir, 1), nil
	}
	return input, nil
}
