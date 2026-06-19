package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type secretExecutor struct {
	namespace string
	name      string
	filename  string
}

func newSecretExecutor(namespace, name, filename string) *secretExecutor {
	return &secretExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *secretExecutor) GetName() string {
	return exe.name
}

func (exe *secretExecutor) Execute() error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get secret from Controller
	secret, err := clt.GetSecret(exe.name)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.SecretKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: rsc.Secret{
			Type: secret.Type,
		},
		Data: secret.Data,
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
