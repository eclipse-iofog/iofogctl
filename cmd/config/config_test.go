package config

import (
	"io/ioutil"
	"testing"
)

var yaml = []byte(`
namespaces:
  - name: first
    controller:
        name: controller1
        user: root
    agents:
      - name: agent1
        user: root
      - name: agent2
        user: root
  - name: second
    controller:
        name: controller2
        user: root
    agents:
      - name: agent3
        user: root
      - name: agent4
        user: root
`)
var filename = "/tmp/cli.yml"
func init() {
	err := ioutil.WriteFile(filename, yaml, 0644)
	if err != nil {
		panic(err)
	}
}

func TestNamespaces(t *testing.T) {
	manager := NewManager(filename)
	namespaces := manager.GetNamespaces()
	if len(namespaces) != 2 {
		t.Errorf("Incorrect number of namespaces: %d", len(namespaces))
	}

	expectedNamespaceNames := [2]string{ "first", "second" }
		for idx, nsName := range expectedNamespaceNames {
		if namespaces[idx].Name != nsName {
			t.Errorf("Namespaces %d incorrect. Expected: %s, Found: %s", idx, namespaces[idx].Name, nsName)
		}
	}
}

func TestControllers(t *testing.T) {

}