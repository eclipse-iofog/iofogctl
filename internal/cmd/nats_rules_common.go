package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func findNatsRuleByName(rules []client.NatsRuleInfo, name string) (client.NatsRuleInfo, error) {
	for _, rule := range rules {
		if rule.Name == name {
			return rule, nil
		}
	}
	return client.NatsRuleInfo{}, util.NewNotFoundError(fmt.Sprintf("NATS rule not found: %s", name))
}

func buildNatsRuleManifest(kind string, rule client.NatsRuleInfo) map[string]interface{} {
	b, err := json.Marshal(rule)
	util.Check(err)

	spec := map[string]interface{}{}
	err = json.Unmarshal(b, &spec)
	util.Check(err)

	// Keep rule manifests Controller-compatible: identity belongs in metadata.
	delete(spec, "id")
	delete(spec, "name")
	delete(spec, "isSystem")

	return map[string]interface{}{
		"apiVersion": util.GetCliApiVersion(),
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": rule.Name,
		},
		"spec": spec,
	}
}
