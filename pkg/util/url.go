package util

import (
	"fmt"
	"net/url"
)

func GetBaseURL(controllerEndpoint string) (*url.URL, error) {
	u, err := url.Parse(controllerEndpoint)
	if err != nil || u.Host == "" {
		// Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		u, err = url.Parse("//" + controllerEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Controller URL (%s): %s", controllerEndpoint, err.Error())
		}
	}

	// Default protocol
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	// Default path
	if u.Path == "" {
		u.Path = "api/v3"
	}
	u.RawQuery = ""
	u.Fragment = ""

	return u, nil
}
