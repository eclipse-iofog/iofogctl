package trust

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// NormalizeTrustCA validates controller trust material from --ca (PEM file) or --ca-b64.
// Returns base64-encoded PEM suitable for spec.ca storage, or empty when neither input is set.
func NormalizeTrustCA(caFile, caB64 string) (string, error) {
	hasFile := strings.TrimSpace(caFile) != ""
	hasB64 := strings.TrimSpace(caB64) != ""
	if hasFile && hasB64 {
		return "", util.NewInputError("Cannot use both --ca and --ca-b64")
	}
	if !hasFile && !hasB64 {
		return "", nil
	}

	var pem []byte
	var err error
	if hasFile {
		pem, err = util.ReadUserFile(caFile)
		if err != nil {
			return "", fmt.Errorf("read CA file: %w", err)
		}
	} else {
		pem, err = base64.StdEncoding.DecodeString(strings.TrimSpace(caB64))
		if err != nil {
			return "", fmt.Errorf("decode CA base64: %w", err)
		}
	}

	if _, err := TransportFromPEM(pem); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pem), nil
}
