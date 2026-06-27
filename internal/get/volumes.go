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

func generateVolumeOutput(namespace string) ([][]string, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	volumes := ns.GetVolumes()

	headers := []string{"VOLUME", "SOURCE", "DESTINATION", "PERMISSIONS", "AGENTS"}
	table := [][]string{headers}

	for _, volume := range volumes {
		agentList := ""
		for idx, agent := range volume.Agents {
			separator := ", "
			if idx == 0 {
				separator = ""
			}
			agentList = agentList + separator + agent
		}
		table = append(table, []string{
			volume.Name,
			volume.Source,
			volume.Destination,
			volume.Permissions,
			agentList,
		})
	}

	return table, nil
}
