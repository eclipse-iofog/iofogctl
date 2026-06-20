package describe

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"gopkg.in/yaml.v2"
)

type configMapExecutor struct {
	namespace string
	name      string
	filename  string
}

func newConfigMapExecutor(namespace, name, filename string) *configMapExecutor {
	return &configMapExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *configMapExecutor) GetName() string {
	return exe.name
}

// Prints the ConfigMap with proper literal block handling
func printConfigMapWithLiteralStrings(header config.Header, writer io.Writer) error {
	// Marshal the header without data first
	headerWithoutData := config.Header{
		APIVersion: header.APIVersion,
		Kind:       header.Kind,
		Metadata:   header.Metadata,
		Spec:       header.Spec,
	}

	data, err := yaml.Marshal(headerWithoutData)
	if err != nil {
		return err
	}

	// Convert to string and remove the closing brace
	yamlStr := string(data)
	yamlStr = strings.TrimSuffix(yamlStr, "\n")

	// Write the header without data
	_, err = writer.Write([]byte(yamlStr))
	if err != nil {
		return err
	}

	// Add the data section with proper literal block formatting
	if dataMap, ok := header.Data.(map[string]string); ok {
		_, err = writer.Write([]byte("\ndata:\n"))
		if err != nil {
			return err
		}

		for key, value := range dataMap {
			if strings.Contains(value, "\n") {
				// Use literal block scalar for multi-line strings
				_, err = fmt.Fprintf(writer, "  %s: |\n", key)
				if err != nil {
					return err
				}

				// Split by newlines and add proper indentation
				lines := strings.Split(value, "\n")
				for _, line := range lines {
					_, err = fmt.Fprintf(writer, "    %s\n", line)
					if err != nil {
						return err
					}
				}
			} else {
				// Regular string
				_, err = fmt.Fprintf(writer, "  %s: %s\n", key, value)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (exe *configMapExecutor) Execute() error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get secret from Controller
	configMap, err := clt.GetConfigMap(exe.name)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ConfigMapKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: rsc.ConfigMap{
			Immutable: configMap.Immutable,
		},
		Data: configMap.Data,
	}

	if exe.filename == "" {
		if err := printConfigMapWithLiteralStrings(header, os.Stdout); err != nil {
			return err
		}
	} else {
		f, err := os.Create(exe.filename)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := printConfigMapWithLiteralStrings(header, f); err != nil {
			return err
		}
	}
	return nil
}
