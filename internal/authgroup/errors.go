package authgroup

import (
	"errors"
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

// MapError translates Controller auth group API errors into CLI-friendly messages.
func MapError(name string, err error) error {
	if err == nil {
		return nil
	}

	var notFound *client.NotFoundError
	if errors.As(err, &notFound) {
		if name != "" {
			return fmt.Errorf("auth group %q not found", name)
		}
		return fmt.Errorf("auth group not found")
	}

	var httpErr *client.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.Code {
		case 501:
			return fmt.Errorf("auth groups require embedded auth mode")
		case 409:
			return fmt.Errorf("auth group name conflict: %w", err)
		case 403:
			return fmt.Errorf("operation not allowed: %w", err)
		case 400:
			return fmt.Errorf("validation error: %w", err)
		case 401:
			return fmt.Errorf("authentication required; reconnect with updated credentials")
		}
	}

	return err
}
