package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type roleExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRoleExecutor(namespace, name, filename string) *roleExecutor {
	return &roleExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *roleExecutor) GetName() string {
	return exe.name
}

func (exe *roleExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	role, err := clt.GetRole(exe.name)
	if err != nil {
		return err
	}

	return printDescribe(exe.filename, roleDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Rules: convertRules(role.Rules),
	})
}

// roleDocument matches Controller-style deploy YAML: rules at the document root.
type roleDocument struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       config.Kind           `yaml:"kind"`
	Metadata   config.HeaderMetadata `yaml:"metadata"`
	Rules      []rsc.RBACRule        `yaml:"rules,omitempty"`
}

func convertRules(rules []client.RBACRule) []rsc.RBACRule {
	out := make([]rsc.RBACRule, len(rules))
	for i, r := range rules {
		out[i] = rsc.RBACRule{
			APIGroups:     r.APIGroups,
			Resources:     r.Resources,
			Verbs:         r.Verbs,
			ResourceNames: r.ResourceNames,
		}
	}
	return out
}
