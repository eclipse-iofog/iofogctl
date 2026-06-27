package describe

import (
	"strconv"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type registryExecutor struct {
	namespace string
	id        int
	filename  string
}

func newRegistryExecutor(namespace, name, filename string) (*registryExecutor, error) {
	a := &registryExecutor{}
	a.namespace = namespace
	id, err := strconv.Atoi(name)
	if err != nil {
		return nil, err
	}
	a.id = id
	a.filename = filename
	return a, nil
}

func (exe *registryExecutor) GetName() string {
	return strconv.Itoa(exe.id)
}

func (exe *registryExecutor) Execute() error {
	// Connect to controller
	ctrl, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	r, err := ctrl.GetRegistry(exe.id)
	if err != nil {
		return err
	}

	private := !r.IsPublic
	registry := rsc.Registry{
		URL:      &r.URL,
		ID:       r.ID,
		Private:  &private,
		Username: &r.Username,
		Email:    &r.Email,
		Password: &r.Password,
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RegistryKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
		},
		Spec: registry,
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
