package validate

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type deployControllerRef struct {
	name string
	host string
}

type deployDocMetadata struct {
	Name string `yaml:"name"`
}

type deployDocHeader struct {
	Kind     config.Kind            `yaml:"kind"`
	Metadata deployDocMetadata      `yaml:"metadata"`
	Spec     map[string]interface{} `yaml:"spec"`
}

// RemoteControllerDeploy rejects same-file ControlPlane controllers[] entries that
// collide by name or host with standalone Controller documents.
func RemoteControllerDeploy(inputFile string) error {
	yamlFile, err := util.ReadUserFile(inputFile)
	if err != nil {
		return err
	}

	var cpControllers []deployControllerRef
	var standaloneControllers []deployControllerRef

	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)

	var doc deployDocHeader
	decodeErr := dec.Decode(&doc)
	for ; !errors.Is(decodeErr, io.EOF); decodeErr = dec.Decode(&doc) {
		if decodeErr != nil {
			return decodeErr
		}
		kind := normalizeDeployKind(doc.Kind)
		switch kind {
		case config.RemoteControlPlaneKind:
			cpControllers = append(cpControllers, controllersFromCPSpec(doc.Spec)...)
		case config.RemoteControllerKind:
			standaloneControllers = append(standaloneControllers, deployControllerRef{
				name: doc.Metadata.Name,
				host: stringFromMap(doc.Spec, "host"),
			})
		}
		doc = deployDocHeader{}
	}

	for _, standalone := range standaloneControllers {
		for _, cpCtrl := range cpControllers {
			if standalone.name != "" && standalone.name == cpCtrl.name {
				return util.NewInputError(fmt.Sprintf(
					"controller name %q appears in both ControlPlane spec.controllers and standalone Controller in the same deploy file",
					standalone.name,
				))
			}
			if standalone.host != "" && standalone.host == cpCtrl.host {
				return util.NewInputError(fmt.Sprintf(
					"controller host %q appears in both ControlPlane spec.controllers and standalone Controller in the same deploy file",
					standalone.host,
				))
			}
		}
	}

	return nil
}

func controllersFromCPSpec(spec map[string]interface{}) []deployControllerRef {
	raw, ok := spec["controllers"]
	if !ok {
		return nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	refs := make([]deployControllerRef, 0, len(items))
	for _, item := range items {
		ctrl, ok := item.(map[interface{}]interface{})
		if !ok {
			continue
		}
		refs = append(refs, deployControllerRef{
			name: stringFromMapAny(ctrl, "name"),
			host: stringFromMapAny(ctrl, "host"),
		})
	}
	return refs
}

func stringFromMap(spec map[string]interface{}, key string) string {
	if spec == nil {
		return ""
	}
	value, ok := spec[key]
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func stringFromMapAny(values map[interface{}]interface{}, key string) string {
	value, ok := values[key]
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func normalizeDeployKind(kind config.Kind) config.Kind {
	if kind == "RemoteController" {
		return config.RemoteControllerKind
	}
	return kind
}
