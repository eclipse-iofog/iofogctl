package exec

import (
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

func lookupMicroservice(clt *client.Client, fqName string) (*client.MicroserviceInfo, bool, error) {
	appName, msvcName, err := clientutil.ParseFQName(fqName, "Microservice")
	if err != nil {
		return nil, false, err
	}

	msvc, err := clt.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid application id") {
			msvc, err = clt.GetSystemMicroserviceByName(appName, msvcName)
			if err != nil {
				return nil, false, err
			}
			return msvc, true, nil
		}
		return nil, false, err
	}

	return msvc, false, nil
}
