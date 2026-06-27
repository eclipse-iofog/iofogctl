package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe *applicationExecutor) initLegacy() (err error) {
	exe.application, err = exe.client.GetApplicationByName(exe.name)
	if err != nil {
		return
	}
	msvcListResponse, err := exe.client.GetMicroservicesByApplication(exe.application.Name)
	if err != nil {
		return
	}

	// Filter system microservices
	for idx := range msvcListResponse.Microservices {
		msvc := &msvcListResponse.Microservices[idx]
		if util.IsSystemMsvc(msvc) {
			continue
		}
		exe.msvcs = append(exe.msvcs, msvc)
	}
	exe.msvcPerID = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(exe.msvcs); i++ {
		exe.msvcPerID[exe.msvcs[i].UUID] = exe.msvcs[i]
	}
	return
}
