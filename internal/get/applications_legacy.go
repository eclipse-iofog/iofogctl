package get

import "github.com/eclipse-iofog/iofogctl/pkg/util"

func (exe *applicationExecutor) initLegacy() (err error) {
	flows, err := exe.client.GetAllFlows()
	if err != nil {
		return
	}
	exe.flows = flows.Flows
	for _, flow := range exe.flows {
		listMsvcs, err := exe.client.GetMicroservicesPerFlow(flow.ID)
		if err != nil {
			return err
		}

		// Filter System microservices
		for idx := range listMsvcs.Microservices {
			msvc := &listMsvcs.Microservices[idx]
			if util.IsSystemMsvc(msvc) {
				continue
			}
			exe.msvcsPerApplication[flow.ID] = append(exe.msvcsPerApplication[flow.ID], msvc)
		}
	}
	return nil
}
