package get

import "github.com/eclipse-iofog/iofogctl/pkg/util"

func (exe *applicationExecutor) initLegacy() (err error) {
	applications, err := exe.client.GetAllApplications()
	if err != nil {
		return
	}
	exe.applications = applications.Applications
	for _, application := range exe.applications {
		listMsvcs, err := exe.client.GetMicroservicesByApplication(application.Name)
		if err != nil {
			return err
		}

		// Filter System microservices
		for idx := range listMsvcs.Microservices {
			msvc := &listMsvcs.Microservices[idx]
			if util.IsSystemMsvc(msvc) {
				continue
			}
			exe.msvcsPerApplication[application.ID] = append(exe.msvcsPerApplication[application.ID], msvc)
		}
	}
	return nil
}
