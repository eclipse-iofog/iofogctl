package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type roleBindingExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRoleBindingExecutor(namespace, name, filename string) *roleBindingExecutor {
	return &roleBindingExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *roleBindingExecutor) GetName() string {
	return exe.name
}

func (exe *roleBindingExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	binding, err := clt.GetRoleBinding(exe.name)
	if err != nil {
		return err
	}

	return printDescribe(exe.filename, roleBindingDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleBindingKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		RoleRef:  convertRoleRef(binding.RoleRef),
		Subjects: convertSubjects(binding.Subjects),
	})
}

// roleBindingDocument matches Controller-style deploy YAML: roleRef and subjects at the document root.
type roleBindingDocument struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       config.Kind           `yaml:"kind"`
	Metadata   config.HeaderMetadata `yaml:"metadata"`
	RoleRef    rsc.RoleRef           `yaml:"roleRef,omitempty"`
	Subjects   []rsc.Subject         `yaml:"subjects,omitempty"`
}

func convertRoleRef(ref client.RoleRef) rsc.RoleRef {
	return rsc.RoleRef{
		Kind:     ref.Kind,
		Name:     ref.Name,
		APIGroup: ref.APIGroup,
	}
}

func convertSubjects(subjects []client.Subject) []rsc.Subject {
	out := make([]rsc.Subject, len(subjects))
	for i, s := range subjects {
		out[i] = rsc.Subject{
			Kind:     s.Kind,
			Name:     s.Name,
			APIGroup: s.APIGroup,
		}
	}
	return out
}
