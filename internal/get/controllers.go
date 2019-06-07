package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
}

func newControllerExecutor(namespace string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	return c
}

func (exe *controllerExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}

	// Generate table and headers
	table := make([][]string, len(controllers)+1)
	headers := []string{"NAME", "STATUS", "AGE", "UPTIME"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ctrlConfig := range controllers {
		// Instantiate connection to controller
		ctrl := iofog.NewController(ctrlConfig.Endpoint)

		// Ping status
		ctrlStatus, err := ctrl.GetStatus()
		uptime := ""
		status := "Failing"
		if err == nil {
			uptimeSec := ctrlStatus.UptimeMsUTC / int64(1000)
			uptime, err = util.Elapsed(util.FromInt(uptimeSec), util.Now())
			if err != nil {
				return err
			}
			status = ctrlStatus.Status
		}

		// Get age
		age, err := util.Elapsed(ctrlConfig.Created, util.Now())
		if err != nil {
			return err
		}
		row := []string{
			ctrlConfig.Name,
			status,
			age,
			uptime,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
