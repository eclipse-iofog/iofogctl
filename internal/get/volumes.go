package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type volumeExecutor struct {
	namespace string
}

func newVolumeExecutor(namespace string) *volumeExecutor {
	c := &volumeExecutor{}
	c.namespace = namespace
	return c
}

func (exe *volumeExecutor) GetName() string {
	return ""
}

func (exe *volumeExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateVolumeOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateVolumeOutput(namespace string) (table [][]string, err error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}
	// Get volume config details
	volumes := ns.GetVolumes()
	if err != nil {
		return
	}

	// Generate table and headers
	table = make([][]string, len(volumes)+1)
	headers := []string{"VOLUME", "SOURCE", "DESTINATION", "PERMISSIONS", "AGENTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, volume := range volumes {
		// Create list of Agents
		agentList := ""
		for idx, agent := range volume.Agents {
			separator := ", "
			if idx == 0 {
				separator = ""
			}
			agentList = agentList + separator + agent
		}
		// Store values
		row := []string{
			volume.Name,
			volume.Source,
			volume.Destination,
			volume.Permissions,
			agentList,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table, err
}
