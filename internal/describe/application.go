package describe

import (
	"errors"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type applicationExecutor struct {
	namespace   string
	name        string
	filename    string
	application *client.ApplicationInfo
	client      *client.Client
	msvcs       []*client.MicroserviceInfo
	msvcPerID   map[string]*client.MicroserviceInfo
	natsCfg     *client.ApplicationNatsConfig
}

func newApplicationExecutor(namespace, name, filename string) *applicationExecutor {
	a := &applicationExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	application, err := exe.client.GetApplicationByName(exe.name)
	// If not found error, try legacy
	notFoundError := &client.NotFoundError{}
	if errors.As(err, &notFoundError) {
		return exe.initLegacy()
	}
	// Return other errors
	if err != nil {
		return err
	}
	// TODO: Use Application instead of application
	exe.application = &client.ApplicationInfo{
		Name:        application.Name,
		IsActivated: application.IsActivated,
		Description: application.Description,
		IsSystem:    application.IsSystem,
		UserID:      application.UserID,
		ID:          application.ID,
	}
	exe.natsCfg = application.NatsConfig
	msvcListResponse, err := exe.client.GetMicroservicesByApplication(exe.name)
	if err != nil {
		return err
	}

	// Filter system microservices
	for idx := range msvcListResponse.Microservices {
		msvc := &msvcListResponse.Microservices[idx]
		if util.IsSystemMsvc(msvc) {
			continue
		}
		exe.msvcs = append(exe.msvcs, msvc)
	}
	exe.msvcPerID = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(exe.msvcs); i++ {
		exe.msvcPerID[exe.msvcs[i].UUID] = exe.msvcs[i]
	}
	return err
}

func (exe *applicationExecutor) GetName() string {
	return exe.name
}

func (exe *applicationExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}

	yamlMsvcs := []rsc.Microservice{}
	var natsCfg *apps.ApplicationNatsConfig
	for idx := range exe.msvcs {
		yamlMsvc, _, _, err := MapClientMicroserviceToDeployMicroservice(exe.msvcs[idx], exe.client)
		if err != nil {
			return err
		}
		// Remove fields
		yamlMsvc.Application = ""
		yamlMsvcs = append(yamlMsvcs, *yamlMsvc)
	}
	if exe.natsCfg != nil {
		natsCfg = &apps.ApplicationNatsConfig{
			NatsAccess: exe.natsCfg.NatsAccess,
			NatsRule:   exe.natsCfg.NatsRule,
		}
	}

	application := rsc.Application{
		Name:          exe.application.Name,
		Microservices: yamlMsvcs,
		NatsConfig:    natsCfg,
		ID:            exe.application.ID,
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ApplicationKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: application,
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
