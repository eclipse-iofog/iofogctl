package wasm

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

const runtimeClassAPIVersion = "edgelet.iofog.org/v1"

type runtimeClassManifest struct {
	APIVersion string               `yaml:"apiVersion"`
	Kind       string               `yaml:"kind"`
	Metadata   runtimeClassMetadata `yaml:"metadata"`
	Handler    string               `yaml:"handler"`
}

type runtimeClassMetadata struct {
	Name string `yaml:"name"`
}

// RuntimeClassManifest builds a single RuntimeClass YAML document for handler.
func RuntimeClassManifest(handler string) ([]byte, error) {
	handler = strings.TrimSpace(handler)
	if handler == "" {
		return nil, fmt.Errorf("WASM handler is required for RuntimeClass manifest")
	}
	doc := runtimeClassManifest{
		APIVersion: runtimeClassAPIVersion,
		Kind:       "RuntimeClass",
		Metadata: runtimeClassMetadata{
			Name: handler,
		},
		Handler: handler,
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal RuntimeClass for %s: %w", handler, err)
	}
	return data, nil
}

// ConfiguredHandlers returns sorted handler keys from package.wasm.
func ConfiguredHandlers(wasm map[string]Pack) []string {
	if len(wasm) == 0 {
		return nil
	}
	handlers := make([]string, 0, len(wasm))
	for handler := range wasm {
		if strings.TrimSpace(handler) == "" {
			continue
		}
		handlers = append(handlers, handler)
	}
	sort.Strings(handlers)
	return handlers
}
