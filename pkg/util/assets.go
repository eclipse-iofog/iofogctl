package util

import (
	"fmt"

	cliassets "github.com/eclipse-iofog/iofogctl/assets"
)

func GetStaticFile(filename string) (string, error) {
	fileContent, err := cliassets.FS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("could not load static file %s: %s", filename, err.Error())
	}
	return string(fileContent), nil
}
