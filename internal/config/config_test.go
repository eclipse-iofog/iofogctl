/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package config

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
)

var configData = []byte(`kind: IofogctlConfig
apiVersion: iofogctl/v1
spec:
  defaultNamespace: default
`)

var defaultData = []byte(`kind: IofogctlNamespace
apiVersion: iofogctl/v1
metadata:
  name: default
spec:
  name: default`)

var firstData = []byte(`kind: IofogctlNamespace
apiVersion: iofogctl/v1
metadata:
  name: first
spec:
  name: first
  controlPlane:
    controllers:
    - name: controller0
      ssh:
        user: root0
  agents:
  - name: agent0
    ssh:
      user: root0
  - name: agent1
    ssh:
      user: root1
  - name: agent2
    ssh:
      user: root2
`)

var secondData = []byte(`kind: IofogctlNamespace
apiVersion: iofogctl/v1
metadata:
  name: second
spec:
  name: second
  controlPlane:
    controllers:
    - name: controller1
      ssh:
        user: root1
  agents:
  - name: agent1
    ssh:
      user: root1
  - name: agent2
    ssh:
      user: root2
`)

func init() {
	testConfigFolder := "/tmp/"
	testConfigFilename := "/tmp/config.yaml"
	os.RemoveAll("/tmp/namespaces")
	if err := os.MkdirAll(path.Join(testConfigFolder, "namespaces"), 0755); err != nil {
		panic(err)
	}
	err := ioutil.WriteFile(testConfigFilename, configData, 0644)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path.Join(testConfigFolder, "namespaces", "first.yaml"), firstData, 0644)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)

	err = ioutil.WriteFile(path.Join(testConfigFolder, "namespaces", "second.yaml"), secondData, 0644)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)

	err = ioutil.WriteFile(path.Join(testConfigFolder, "namespaces", "default.yaml"), defaultData, 0644)
	if err != nil {
		panic(err)
	}

	Init(testConfigFolder)
}

func TestDelete(t *testing.T) {
	DeleteAgent("first", "agent2")
}

func TestReadingNamespaces(t *testing.T) {
	// Test all namespace queries
	namespaces := GetNamespaces()
	if len(namespaces) != 3 {
		t.Errorf("Incorrect number of namespaces: %d", len(namespaces))
	}
	expectedNamespaceNames := []string{"first", "second", "default"}
	for _, nsName := range expectedNamespaceNames {
		found := false

		for idx := range namespaces {
			if namespaces[idx] == nsName {
				found = true
			}
		}
		if !found {
			t.Errorf("Could not find expected namespace " + nsName)
		}

		// Test single namespace queries
		singleNamespace, err := GetNamespace(nsName)
		if err != nil {
			t.Errorf("[%s] Error getting namespace. Error: %s", nsName, err.Error())
		}
		if singleNamespace.Name != nsName {
			t.Errorf("[%v] Error getting namespace. Expected: %s, Found: %s", singleNamespace, nsName, singleNamespace.Name)
		}
	}

	// Negative tests
	_, err := GetNamespace("falsename")
	if err == nil {
		t.Errorf("Expected error when requested non-existing namespace")
	}
}

func TestReadingControllers(t *testing.T) {
	for _, ns := range GetNamespaces() {
		// Test bulk Controller queries
		ctrls, err := GetControllers(ns)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		for ctrlIdx, ctrl := range ctrls {
			idx := ctrlIdx
			if ns == "second" {
				idx += 1
			}
			expectedName := "controller" + strconv.Itoa(idx)
			if ctrl.Name != expectedName {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedName, ctrl.Name)
			}
			expectedUser := "root" + strconv.Itoa(idx)
			if ctrl.SSH.User != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedUser, ctrl.SSH.User)
			}

			// Test single Controller queries
			singleCtrl, err := GetController(ns, expectedName)
			if err != nil {
				t.Errorf("Error getting single Controller: %s", err.Error())
			}
			if singleCtrl.Name != expectedName {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedName, singleCtrl.Name)
			}
			if singleCtrl.SSH.User != expectedUser {
				t.Errorf("Error in Controller name. Expected %s, Found: %s", expectedUser, singleCtrl.SSH.User)
			}
		}
	}
}

func TesReadingtAgents(t *testing.T) {
	for _, ns := range GetNamespaces() {
		// Test bulk Agent queries
		agents, err := GetAgents(ns)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		for agentIdx, agent := range agents {
			idx := agentIdx
			if ns == "second" {
				idx += 1
			}
			expectedName := "agent" + strconv.Itoa(idx)
			if agent.Name != expectedName {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, agent.Name)
			}
			expectedUser := "root" + strconv.Itoa(idx)
			if agent.SSH.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, agent.SSH.User)
			}

			// Test single Agent queries
			singleAgent, err := GetAgent(ns, expectedName)
			if err != nil {
				t.Errorf("Error getting single Agent: %s", err.Error())
			}
			if singleAgent.Name != expectedName {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedName, singleAgent.Name)
			}
			if singleAgent.SSH.User != expectedUser {
				t.Errorf("Error in Agent name. Expected %s, Found: %s", expectedUser, singleAgent.SSH.User)
			}
		}
	}
}

var writeNamespace = "write_namespace"

func TestWritingNamespace(t *testing.T) {
	if err := AddNamespace(writeNamespace, ""); err != nil {
		t.Errorf("Error adding write namespace: %s", err.Error())
	}
	if _, err := GetNamespace(writeNamespace); err != nil {
		t.Errorf("Error getting write namespace: %s", err.Error())
	}
}

func compareControllers(lhs, rhs Controller) bool {
	equal := (lhs.Created == rhs.Created)
	equal = equal && (lhs.Endpoint == rhs.Endpoint)
	equal = equal && (lhs.Host == lhs.Host)
	equal = equal && (lhs.SSH.KeyFile == rhs.SSH.KeyFile)
	equal = equal && (lhs.Name == rhs.Name)
	equal = equal && (lhs.SSH.User == lhs.SSH.User)

	return equal
}
func TestWritingController(t *testing.T) {
	ctrl := Controller{
		Created:  "Now",
		Endpoint: "localhost:" + iofog.ControllerPortString,
		Host:     "localhost",
		SSH: SSH{
			User:    "Kubert",
			KeyFile: "~/.key/file",
		},
		Name: "Hubert",
	}
	if err := AddController(writeNamespace, ctrl); err != nil {
		t.Errorf("Error Creating controller in write namespace: %s", err.Error())
	}
	ctrlTwo := ctrl
	ctrlTwo.Name = "ctrlTwo"
	if err := AddController(writeNamespace, ctrlTwo); err != nil {
		t.Errorf("Error Creating controller in write namespace: %s", err.Error())
	}

	readCtrl, err := GetController(writeNamespace, ctrl.Name)
	if err != nil {
		t.Errorf("Error reading Controller from write namespace: %s", err.Error())
	}
	if !compareControllers(ctrl, readCtrl) {
		t.Error("Written Controller is not identical to read Controller")
	}
	if compareControllers(ctrlTwo, readCtrl) {
		t.Error("Expected different Controllers to not be identical")
	}

	ctrlTwo.Host = "changed"
	if err = UpdateController(writeNamespace, ctrlTwo); err != nil {
		t.Errorf("Error updating Controller in write namespace: %s", err.Error())
	}

	readCtrl, err = GetController(writeNamespace, ctrlTwo.Name)
	if err != nil {
		t.Errorf("Error reading Controller from write namespace: %s", err.Error())
	}
	if !compareControllers(ctrlTwo, readCtrl) {
		t.Error("Written Controller is not identical to read Controller")
	}
	if compareControllers(ctrl, readCtrl) {
		t.Error("Expected different Controllers to not be identical")
	}
}

func compareAgents(lhs, rhs Agent) bool {
	equal := (lhs.Created == rhs.Created)
	equal = equal && (lhs.Host == rhs.Host)
	equal = equal && (lhs.Container.Image == rhs.Container.Image)
	equal = equal && (lhs.SSH.KeyFile == rhs.SSH.KeyFile)
	equal = equal && (lhs.Name == rhs.Name)
	equal = equal && (lhs.SSH.Port == rhs.SSH.Port)
	equal = equal && (lhs.UUID == rhs.UUID)
	equal = equal && (lhs.SSH.User == rhs.SSH.User)

	return equal
}

func TestWritingAgent(t *testing.T) {
	agent := Agent{
		Created: "Now",
		Host:    "localhost",
		SSH: SSH{
			User:    "Kubert",
			KeyFile: "~/.key/file",
		},
		Name: "Hubert",
	}
	if err := AddAgent(writeNamespace, agent); err != nil {
		t.Errorf("Error Creating Agent in write namespace: %s", err.Error())
	}
	agentTwo := agent
	agentTwo.Name = "agentTwo"
	if err := AddAgent(writeNamespace, agentTwo); err != nil {
		t.Errorf("Error Creating Agent in write namespace: %s", err.Error())
	}

	readAgent, err := GetAgent(writeNamespace, agent.Name)
	if err != nil {
		t.Errorf("Error reading Agent from write namespace: %s", err.Error())
	}
	if !compareAgents(agent, readAgent) {
		t.Error("Written Agent is not identical to read Agent")
	}
	if compareAgents(agentTwo, readAgent) {
		t.Error("Expected different Agents to not be identical")
	}

	agentTwo.Host = "changed"
	if err = UpdateAgent(writeNamespace, agentTwo); err != nil {
		t.Errorf("Error updating Agent in write namespace: %s", err.Error())
	}

	readAgent, err = GetAgent(writeNamespace, agentTwo.Name)
	if err != nil {
		t.Errorf("Error reading Agent from write namespace: %s", err.Error())
	}
	if !compareAgents(agentTwo, readAgent) {
		t.Error("Written Agent is not identical to read Agent")
	}
	if compareAgents(agent, readAgent) {
		t.Error("Expected different Agent to not be identical")
	}
}
