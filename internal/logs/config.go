package logs

import (
	"fmt"
	"net/url"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// LogTailConfig holds configuration for log tailing
type LogTailConfig struct {
	Tail   int    // Number of lines to tail (default: 100, range: 1-10000)
	Follow bool   // Whether to follow logs (default: true)
	Since  string // Start time in ISO 8601 format (optional)
	Until  string // End time in ISO 8601 format (optional)
}

// DefaultLogTailConfig returns a LogTailConfig with default values
func DefaultLogTailConfig() *LogTailConfig {
	return &LogTailConfig{
		Tail:   100,
		Follow: true,
	}
}

// Validate validates the LogTailConfig
func (c *LogTailConfig) Validate() error {
	// Validate tail range
	if c.Tail < 1 || c.Tail > 10000 {
		return util.NewInputError(fmt.Sprintf("tail must be between 1 and 10000, got %d", c.Tail))
	}

	// Validate ISO 8601 format for since if provided
	if c.Since != "" {
		if err := validateISO8601(c.Since); err != nil {
			return util.NewInputError(fmt.Sprintf("invalid since format: %v (expected ISO 8601 format)", err))
		}
	}

	// Validate ISO 8601 format for until if provided
	if c.Until != "" {
		if err := validateISO8601(c.Until); err != nil {
			return util.NewInputError(fmt.Sprintf("invalid until format: %v (expected ISO 8601 format)", err))
		}
	}

	return nil
}

// BuildQueryString builds a query string from the LogTailConfig
func (c *LogTailConfig) BuildQueryString() string {
	values := url.Values{}

	// Always include tail (default is 100)
	values.Set("tail", fmt.Sprintf("%d", c.Tail))

	// Always include follow (default is true)
	values.Set("follow", fmt.Sprintf("%t", c.Follow))

	// Include since if provided
	if c.Since != "" {
		values.Set("since", c.Since)
	}

	// Include until if provided
	if c.Until != "" {
		values.Set("until", c.Until)
	}

	return values.Encode()
}

// validateISO8601 validates that a string is in ISO 8601 format
func validateISO8601(dateStr string) error {
	// Try parsing with RFC3339 format (ISO 8601 compatible)
	_, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try parsing with RFC3339Nano format
		_, err = time.Parse(time.RFC3339Nano, dateStr)
		if err != nil {
			return fmt.Errorf("invalid ISO 8601 format: %w", err)
		}
	}
	return nil
}
