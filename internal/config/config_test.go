package config

import (
	"io/ioutil"
	"strconv"
	"testing"
)

var testData = []byte(`
namespaces:
- name: first
  controllers:
  - name: controller0
    user: root0
  agents:
  - name: agent0
    user: root0
  - name: agent1
    user: root1
  - name: agent2
    user: root2
- name: second
  controllers:
  - name: controller1
    user: root1
  agents:
  - name: agent1
    user: root1
  - name: agent2
    user: root2
`)

func init() {
	testConfigFilename := "/tmp/cli.yml"
	err := ioutil.WriteFile(testConfigFilename, testData, 0644)
	if err != nil {
		panic(err)
	}
	Init(testConfigFilename)
}

func TestDelete(t *testing.T) {
	DeleteAgent("first", "agent2")
}

func TestNamespaces(t *testing.T) {
	// Test all namespace queries
	namespaces := GetNamespaces()
	if len(namespaces) != 2 {
		t.Errorf("Incorrect number of namespaces: %d", len(namespaces))
	}
	expectedNamespaceNames := [2]string{"first", "second"}
	for idx, nsName := range expectedNamespaceNames {
		if namespaces[idx].Name != nsName {
			t.Errorf("Namespaces %d incorrect. Expected: %s, Found: %s", idx, namespaces[idx].Name, nsName)
		}

		// Test single namespace queries
		singleNamespace, err := GetNamespace(nsName)
		if err != nil {
			t.Errorf("Error getting namespace. Error: %s", err.Error())
		}
		if singleNamespace.Name != nsName {
			t.Errorf("Error getting namespace. Expected: %s, Found: %s", nsName, singleNamespace.Name)
		}
	}

	// Negative tests
	_, err := GetNamespace("falsename")
	if err == nil {
		t.Errorf("Expected error when requested non-existing namespace")
	}
}

func TestControllers(t *testing.T) {
	for nsIdx, ns := range GetNamespaces() {
		// Test bulk Controller queries
		ctrls, err := GetControllers(ns.Name)
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
			singleCtrl, err := GetController(ns.Name, expectedName)
			if err != nil {
				t.Errorf("Error getting single Controller: %s", err.Error())
			}
			if singleCtrl.Name != expectedName {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedName, singleCtrl.Name)
			}
			if singleCtrl.User != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedUser, singleCtrl.User)
			}
		}
	}
}

func TestAgents(t *testing.T) {
	for nsIdx, ns := range GetNamespaces() {
		// Test bulk Agent queries
		agents, err := GetAgents(ns.Name)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		for agentIdx, agent := range agents {
			idx := nsIdx + agentIdx
			expectedName := "agent" + strconv.Itoa(idx)
			if agent.Name != expectedName {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, agent.Name)
			}
			expectedUser := "root" + strconv.Itoa(idx)
			if agent.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, agent.User)
			}

			// Test single Agent queries
			singleAgent, err := GetAgent(ns.Name, expectedName)
			if err != nil {
				t.Errorf("Error getting single Agent: %s", err.Error())
			}
			if singleAgent.Name != expectedName {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, singleAgent.Name)
			}
			if singleAgent.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, singleAgent.User)
			}
		}
	}
}
