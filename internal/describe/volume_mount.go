package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type volumeMountExecutor struct {
	namespace string
	name      string
	filename  string
}

func newVolumeMountExecutor(namespace, name, filename string) *volumeMountExecutor {
	return &volumeMountExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *volumeMountExecutor) GetName() string {
	return exe.name
}

func (exe *volumeMountExecutor) Execute() error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Volume Mount from Controller
	volumeMount, err := clt.GetVolumeMount(exe.name)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.VolumeMountKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: rsc.VolumeMount{
			Name:          volumeMount.Name,
			UUID:          volumeMount.UUID,
			ConfigMapName: volumeMount.ConfigMapName,
			SecretName:    volumeMount.SecretName,
			Version:       volumeMount.Version,
		},
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
