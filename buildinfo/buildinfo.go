// Package buildinfo hosts build info variables populated via ldflags.
package buildinfo

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// DefaultVersion is the default value for Version if not
// set via ldflags.
const (
	DefaultVersion = "v0.0.0-dev"
	// RFC3339Milli is an RFC3339 format with millisecond precision.
	RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"
	// ISO8601 is similar to RFC3339Milli, but doesn't have the colon
	// in the timezone offset.
	ISO8601 = "2006-01-02T15:04:05.000Z0700"

	// DateOnly is a date-only format.
	DateOnly = "2006-01-02"
)

var (
	// Version is the build version. If not set at build time via
	// ldflags, Version takes the value of DefaultVersion.
	Version = DefaultVersion

	// Commit is the commit hash.
	Commit string

	// Timestamp is the timestamp of when the cli was built.
	Timestamp string
)

// BuildInfo encapsulates Version, Commit and Timestamp.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// String returns a string representation of BuildInfo.
func (bi BuildInfo) String() string {
	s := bi.Version
	if bi.Commit != "" {
		s += " " + bi.Commit
	}
	if bi.Timestamp != "" {
		s += " " + bi.Timestamp
	}
	return s
}

// Info returns BuildInfo.
func Info() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		Timestamp: Timestamp,
	}
}

func init() { //nolint:gochecknoinits
	if strings.HasSuffix(Version, "~dev") {
		Version = strings.Replace(Version, "~dev", "-dev", 1)
	}

	if Version != "" && !semver.IsValid(Version) {
		// We want to panic here because it is a pipeline/build failure
		// to have an invalid non-empty Version.
		panic(fmt.Sprintf("Invalid BuildInfo.Version value: %q", Version))
	}

	if Timestamp != "" {
		// Make sure Timestamp is normalized
		t := TimestampToRFC3339(Timestamp)
		if t != "" {
			Timestamp = t
		}
	}
}

// IsDefaultVersion returns true if Version is empty or DefaultVersion.
func IsDefaultVersion() bool {
	return Version == "" || Version == DefaultVersion
}

// TimestampToRFC3339 takes a RFC3339Milli, ISO8601 or RFC3339
// timestamp, and returns RFC3339. That is, the milliseconds are dropped.
// On error, the empty string is returned.
func TimestampToRFC3339(s string) string {
	t, err := ParseTimestampUTC(s)
	if err != nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// ParseTimestampUTC is the counterpart of TimestampUTC. It attempts
// to parse s first in RFC3339Milli, then time.RFC3339 format, falling
// back to the subtly different ISO8601 format.
func ParseTimestampUTC(s string) (time.Time, error) {
	t, err := time.Parse(RFC3339Milli, s)
	if err == nil {
		return t.UTC(), nil
	}

	// Fallback to RFC3339
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return t.UTC(), nil
	}

	// Fallback to ISO8601
	t, err = time.Parse(ISO8601, s)
	if err == nil {
		return t.UTC(), nil
	}

	return t.UTC(), err
}
