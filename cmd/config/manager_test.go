package config

import (
	"strconv"
	"io/ioutil"
	"testing"
)

var testData = []byte(`
namespaces:
  - name: first
    controller:
        name: controller0
        user: root0
    agents:
      - name: agent0
        user: root0
      - name: agent1
        user: root1
  - name: second
    controller:
        name: controller1
        user: root1
    agents:
      - name: agent1
        user: root1
      - name: agent2
        user: root2
`)
var filename = "/tmp/cli.yml"
func init() {
	err := ioutil.WriteFile(filename, testData, 0644)
	if err != nil {
		panic(err)
	}
}

func TestNamespaces(t *testing.T) {
	manager := NewManager(filename)
	
	// Test all namespace queries
	namespaces := manager.GetNamespaces()
	if len(namespaces) != 2 {
		t.Errorf("Incorrect number of namespaces: %d", len(namespaces))
	}
	expectedNamespaceNames := [2]string{ "first", "second" }
	for idx, nsName := range expectedNamespaceNames {
		if namespaces[idx].Name != nsName {
			t.Errorf("Namespaces %d incorrect. Expected: %s, Found: %s", idx, namespaces[idx].Name, nsName)
		}

		// Test single namespace queries
		singleNamespace, err := manager.GetNamespace(nsName)
		if err != nil {
			t.Errorf("Error getting namespace. Error: %s", err.Error())
		}
		if singleNamespace.Name != nsName {
			t.Errorf("Error getting namespace. Expected: %s, Found: %s", nsName, singleNamespace.Name)
		}
	}

	// Negative tests
	_, err := manager.GetNamespace("falsename")
	if err == nil {
		t.Errorf("Expected error when requested non-existing namespace")
	}
}

func TestControllers(t *testing.T) {
	manager := NewManager(filename)
	for nsIdx, ns := range manager.GetNamespaces() {
		// Test bulk Controller queries
		ctrls, err := manager.GetControllers(ns.Name)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		for ctrlIdx, ctrl := range ctrls {
			idx := nsIdx + ctrlIdx
			expectedName := "controller" + strconv.Itoa(idx)
			if ctrl.Name != expectedName {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedName, ctrl.Name)
			}
			expectedUser := "root" + strconv.Itoa(idx)
			if ctrl.User != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedUser, ctrl.User)
			}

			// Test single Controller queries
			singleCtrl, err := manager.GetController(ns.Name, expectedName)
			if err != nil {
				t.Errorf("Error getting single Controller: %s", err.Error())
			}
			if singleCtrl.Name != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedName, singleCtrl.Name)
			}
			if singleCtrl.User != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedUser, singleCtrl.User)
			}
		}
	}
}

func TestAgents(t *testing.T) {
	manager := NewManager(filename)
	for nsIdx, ns := range manager.GetNamespaces() {
		// Test bulk Agent queries
		agents, err := manager.GetAgents(ns.Name)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		for agentIdx, agent := range agents {
			idx := nsIdx + agentIdx
			expectedName := "controller" + strconv.Itoa(idx)
			if agent.Name != expectedName {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, agent.Name)
			}
			expectedUser := "root" + strconv.Itoa(idx)
			if agent.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, agent.User)
			}

			// Test single Agent queries
			singleAgent, err := manager.GetAgent(ns.Name, expectedName)
			if err != nil {
				t.Errorf("Error getting single Agent: %s", err.Error())
			}
			if singleAgent.Name != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, singleAgent.Name)
			}
			if singleAgent.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, singleAgent.User)
			}
		}
	}
}

func TestDelete(t *testing.T){
	manager := NewManager(filename)
	manager.DeleteAgent("first", "agent1")
}