package describe

import (
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type serviceAccountExecutor struct {
	namespace string
	appName   string
	name      string
	filename  string
}

// parseServiceAccountName returns (appName, name). Name can be "appName/name" or just "name" (appName then empty for list lookup).
func parseServiceAccountName(arg string) (appName, name string) {
	if idx := strings.Index(arg, "/"); idx >= 0 {
		return arg[:idx], arg[idx+1:]
	}
	return "", arg
}

func newServiceAccountExecutor(namespace, name, filename string) *serviceAccountExecutor {
	appName, saName := parseServiceAccountName(name)
	return &serviceAccountExecutor{
		namespace: namespace,
		appName:   appName,
		name:      saName,
		filename:  filename,
	}
}

func (exe *serviceAccountExecutor) GetName() string {
	if exe.appName != "" {
		return exe.appName + "/" + exe.name
	}
	return exe.name
}

func (exe *serviceAccountExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if exe.appName == "" {
		return util.NewInputError("ServiceAccount is application-scoped: use APPLICATION_NAME/SERVICE_ACCOUNT_NAME (e.g. myapp/my-sa)")
	}

	sa, err := clt.GetServiceAccount(exe.appName, exe.name)
	if err != nil {
		return err
	}

	return printDescribe(exe.filename, serviceAccountDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ServiceAccountKind,
		Metadata: config.HeaderMetadata{
			Namespace:       exe.namespace,
			Name:            exe.name,
			ApplicationName: sa.ApplicationName,
		},
		RoleRef: convertRoleRef(sa.RoleRef),
	})
}

// serviceAccountDocument matches Controller-style deploy YAML: applicationName on metadata, roleRef at the root.
type serviceAccountDocument struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       config.Kind           `yaml:"kind"`
	Metadata   config.HeaderMetadata `yaml:"metadata"`
	RoleRef    rsc.RoleRef           `yaml:"roleRef,omitempty"`
}
