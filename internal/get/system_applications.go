package get

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type systemApplicationExecutor struct {
	namespace           string
	client              *client.Client
	applications        []client.ApplicationInfo
	msvcsPerApplication map[int][]*client.MicroserviceInfo
}

func newSystemApplicationExecutor(namespace string) *systemApplicationExecutor {
	c := &systemApplicationExecutor{}
	c.namespace = namespace
	c.msvcsPerApplication = make(map[int][]*client.MicroserviceInfo)
	return c
}

func (exe *systemApplicationExecutor) GetName() string {
	return ""
}

func (exe *systemApplicationExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	table := exe.generateSystemApplicationOutput()
	return print(table)
}

func (exe *systemApplicationExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}
	applications, err := exe.client.GetAllSystemApplications()
	// // Try legacy if error is "not found"
	// if _, ok := err.(*client.NotFoundError); ok {
	// 	if err := exe.initLegacy(); err != nil {
	// 		return err
	// 	}
	// 	// Successful legacy
	// 	return nil
	// }
	if err != nil {
		// Return errors that are not "not found"
		return err
	}
	// Execute non-legacy
	// Map applications to application
	// TODO: Use Application instead of application
	exe.applications = []client.ApplicationInfo{}
	for _, application := range applications.Applications {
		exe.applications = append(exe.applications, client.ApplicationInfo{
			Name:        application.Name,
			IsActivated: application.IsActivated,
			Description: application.Description,
			IsSystem:    application.IsSystem,
			UserID:      application.UserID,
			ID:          application.ID,
		})
		listMsvcs, err := exe.client.GetSystemMicroservicesByApplication(application.Name)
		if err != nil {
			return err
		}

		// Filter System microservices
		for idx := range listMsvcs.Microservices {
			msvc := &listMsvcs.Microservices[idx]
			// if util.IsSystemMsvc(msvc) {
			// 	continue
			// }
			exe.msvcsPerApplication[application.ID] = append(exe.msvcsPerApplication[application.ID], msvc)
		}
	}
	return err
}

func (exe *systemApplicationExecutor) generateSystemApplicationOutput() (table [][]string) {
	// Generate table and headers
	table = make([][]string, len(exe.applications)+1)
	headers := []string{"SYS-APPLICATION", "RUNNING", "SYS-MICROSERVICES"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, application := range exe.applications {
		nbMsvcs := len(exe.msvcsPerApplication[application.ID])
		runningMsvcs := 0
		msvcs := ""
		first := true
		for idx := range exe.msvcsPerApplication[application.ID] {
			msvc := exe.msvcsPerApplication[application.ID][idx]
			if first {
				msvcs += msvc.Name
			} else {
				msvcs += fmt.Sprintf(", %s", msvc.Name)
			}
			first = false
			if msvc.Status.Status == "RUNNING" {
				runningMsvcs++
			}
		}

		if nbMsvcs > 5 {
			msvcs = fmt.Sprintf("%d microservices", len(exe.msvcsPerApplication[application.ID]))
		}

		status := fmt.Sprintf("%d/%d", runningMsvcs, nbMsvcs)

		row := []string{
			application.Name,
			status,
			msvcs,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table
}
