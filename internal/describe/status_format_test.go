package describe

import "gopkg.in/yaml.v2"

func statusMap(s yaml.MapSlice) map[string]interface{} {
	m := make(map[string]interface{}, len(s))
	for _, item := range s {
		if k, ok := item.Key.(string); ok {
			m[k] = item.Value
		}
	}
	return m
}

func statusKeys(s yaml.MapSlice) []string {
	keys := make([]string, 0, len(s))
	for _, item := range s {
		if k, ok := item.Key.(string); ok {
			keys = append(keys, k)
		}
	}
	return keys
}
