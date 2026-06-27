package get

import (
	"regexp"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func getAddressAndPort(endpoint, defaultPort string) (addr, port string) {
	// Remove prefix
	regex := regexp.MustCompile("https?://")
	addr = regex.ReplaceAllString(endpoint, "")

	// Get port from address
	port = util.AfterLast(addr, ":")
	if port == "" {
		port = defaultPort
	}

	// Remove port from address
	addr = util.Before(addr, ":")

	return
}
