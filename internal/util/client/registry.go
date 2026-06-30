package client

import (
	"errors"
	"strconv"

	sdkclient "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// ResolveRegistryID parses deploy YAML registry values: built-in names (remote, local) or numeric ids.
func ResolveRegistryID(registry string) (int, error) {
	if registry == "" {
		return 1, nil
	}
	if id, ok := sdkclient.RegistryTypeRegistryTypeIDDict[registry]; ok {
		return id, nil
	}
	id, err := strconv.Atoi(registry)
	if err != nil || id <= 0 {
		return 0, util.NewInputError("Registry must be a valid registry id, 'remote', or 'local'")
	}
	return id, nil
}

// FormatRegistryID formats a controller registry id for CLI output.
func FormatRegistryID(registryID int) string {
	return strconv.Itoa(registryID)
}

// IsClientNotFoundError reports whether err is a controller client not-found error.
func IsClientNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var notFound *sdkclient.NotFoundError
	return errors.As(err, &notFound)
}
