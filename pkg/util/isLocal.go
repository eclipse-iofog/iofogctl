package util

import "regexp"

func IsLocalHost(host string) bool {
	r := regexp.MustCompile("^(http(s){0,1}:\\/\\/){0,1}(localhost|0\\.0\\.0\\.0|127\\.0\\.0\\.1)")
	return r.MatchString(host)
}
