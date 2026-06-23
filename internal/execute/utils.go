package execute

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

// headerDecode accepts both iofogctl-style (spec: ...) and Controller-style RBAC YAML
// (top-level rules, roleRef, subjects) so deploy works with Controller examples.
type headerDecode struct {
	APIVersion string                `yaml:"apiVersion"`
	Kind       config.Kind           `yaml:"kind"`
	Metadata   config.HeaderMetadata `yaml:"metadata"`
	Spec       interface{}           `yaml:"spec,omitempty"`
	Data       interface{}           `yaml:"data,omitempty"`
	Status     interface{}           `yaml:"status,omitempty"`
	Rules      interface{}           `yaml:"rules,omitempty"`
	RoleRef    interface{}           `yaml:"roleRef,omitempty"`
	Subjects   interface{}           `yaml:"subjects,omitempty"`
}

type emptyExecutor struct {
	name string
}

func (exe *emptyExecutor) Execute() error {
	return nil
}
func (exe *emptyExecutor) GetName() string {
	return exe.name
}

// NewEmptyExecutor return an executor that does nothing
func NewEmptyExecutor(name string) Executor {
	return &emptyExecutor{
		name: name,
	}
}

func generateExecutor(header *config.Header, namespace string, deleteNamespace bool, kindHandlers map[config.Kind]func(*KindHandlerOpt) (Executor, error)) (exe Executor, err error) {
	if len(header.Metadata.Namespace) > 0 && namespace != header.Metadata.Namespace {
		msg := "The Namespace provided by the %s named '%s' does not match the Namespace '%s'. You must pass '--namespace %s' to perform this command"
		return nil, util.NewInputError(fmt.Sprintf(msg, header.Kind, header.Metadata.Name, namespace, header.Metadata.Namespace))
	}

	if _, err := config.GetNamespace(namespace); err != nil {
		return nil, err
	}

	if err := config.ValidateHeader(header); err != nil {
		return nil, err
	}

	subYamlBytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return exe, err
	}

	dataYamlBytes, err := yaml.Marshal(header.Data)
	if err != nil {
		return exe, err
	}
	fullYamlBytes, err := yaml.Marshal(header)
	if err != nil {
		return exe, err
	}

	createExecutorFunc, found := kindHandlers[header.Kind]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not handle kind %s. Skipping document\n", header.Kind))
		return nil, nil
	}

	return createExecutorFunc(&KindHandlerOpt{
		Kind:            header.Kind,
		Namespace:       namespace,
		Name:            header.Metadata.Name,
		YAML:            subYamlBytes,
		FullYAML:        fullYamlBytes,
		Data:            dataYamlBytes,
		Tags:            header.Metadata.Tags,
		DeleteNamespace: deleteNamespace,
	})
}

type KindHandlerOpt struct {
	Kind            config.Kind
	Namespace       string
	Name            string
	YAML            []byte
	FullYAML        []byte
	Data            []byte
	Tags            *[]string
	DeleteNamespace bool
}

func GetExecutorsFromYAML(inputFile, namespace string, kindHandlers map[config.Kind]func(*KindHandlerOpt) (Executor, error), deleteNamespace bool) (executorsMap map[config.Kind][]Executor, err error) {
	yamlFile, err := os.ReadFile(inputFile)
	if err != nil {
		return
	}

	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)

	var h headerDecode

	// Generate all executors
	empty := true
	executorsMap = make(map[config.Kind][]Executor)
	decodeErr := dec.Decode(&h)
	for decodeErr == nil {
		header := headerDecodeToHeader(&h)
		exe, err := generateExecutor(header, namespace, deleteNamespace, kindHandlers)
		if err != nil {
			return nil, err
		}
		if exe != nil {
			empty = false
			executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
		}

		// Reset for next document
		h = headerDecode{}

		decodeErr = dec.Decode(&h)
	}
	if !errors.Is(decodeErr, io.EOF) {
		return nil, decodeErr
	}

	if empty {
		err = util.NewInputError("Could not decode any valid resources from input YAML file")
	}

	return executorsMap, err
}

// headerDecodeToHeader converts headerDecode to config.Header, building Spec from
// top-level rules/roleRef/subjects when present (Controller-style RBAC YAML).
func headerDecodeToHeader(h *headerDecode) *config.Header {
	kind := h.Kind
	if kind == "RemoteController" {
		kind = config.RemoteControllerKind
	}

	header := &config.Header{
		APIVersion: h.APIVersion,
		Kind:       kind,
		Metadata:   h.Metadata,
		Spec:       h.Spec,
		Data:       h.Data,
		Status:     h.Status,
	}

	switch kind {
	case config.RoleKind:
		if h.Rules != nil {
			// Controller-style: rules at top level
			header.Spec = map[string]interface{}{
				"name":  h.Metadata.Name,
				"kind":  "Role",
				"rules": h.Rules,
			}
		}
	case config.RoleBindingKind:
		if h.RoleRef != nil || h.Subjects != nil {
			// Controller-style: roleRef and subjects at top level
			spec := map[string]interface{}{"name": h.Metadata.Name}
			if h.RoleRef != nil {
				spec["roleRef"] = h.RoleRef
			}
			if h.Subjects != nil {
				spec["subjects"] = h.Subjects
			}
			header.Spec = spec
		}
	case config.ServiceAccountKind:
		if h.RoleRef != nil {
			// Controller-style: metadata.applicationName and roleRef at top level
			header.Spec = map[string]interface{}{
				"name":            h.Metadata.Name,
				"applicationName": h.Metadata.ApplicationName,
				"roleRef":         h.RoleRef,
			}
		}
	}

	return header
}
