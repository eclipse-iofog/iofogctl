package nodeversion

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

// MapError translates Controller node version command errors into CLI-friendly messages.
func MapError(action string, err error) error {
	if err == nil {
		return nil
	}

	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		return err
	}

	msg := strings.ToLower(httpErr.Error())
	switch {
	case httpErr.Code == 400 && strings.Contains(msg, "invalid_version_command_upgrade"):
		return fmt.Errorf("agent is not ready to upgrade")
	case httpErr.Code == 400 && strings.Contains(msg, "invalid_version_command_rollback"):
		return fmt.Errorf("agent is not ready to rollback")
	case httpErr.Code == 400 && strings.Contains(msg, "validationerror"):
		return fmt.Errorf("invalid semver for %s: %w", action, err)
	case httpErr.Code == 400:
		return fmt.Errorf("%s failed: %w", action, err)
	}

	return err
}
