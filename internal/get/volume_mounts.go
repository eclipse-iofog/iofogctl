package get

import (
	"strconv"

	// "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type volumeMountExecutor struct {
	namespace string
}

func newVolumeMountExecutor(namespace string) *volumeMountExecutor {
	a := &volumeMountExecutor{}
	a.namespace = namespace
	return a
}

func (exe *volumeMountExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateVolumeMountsOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func (exe *volumeMountExecutor) GetName() string {
	return ""
}

func generateVolumeMountsOutput(namespace string) ([][]string, error) {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	volumeMountList, err := clt.ListVolumeMounts()
	if err != nil {
		return nil, err
	}

	// Generate table and headers
	table := make([][]string, len(volumeMountList.VolumeMounts)+1)
	headers := []string{
		"VOLUME MOUNT",
		"CONFIG MAP",
		"SECRET",
		"VERSION",
	}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, volumeMount := range volumeMountList.VolumeMounts {
		// Handle empty values

		configMap := ""
		if volumeMount.ConfigMapName != "" {
			configMap = volumeMount.ConfigMapName
		}

		secret := ""
		if volumeMount.SecretName != "" {
			secret = volumeMount.SecretName
		}

		row := []string{
			volumeMount.Name,
			configMap,
			secret,
			strconv.Itoa(volumeMount.Version),
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table, nil
}
