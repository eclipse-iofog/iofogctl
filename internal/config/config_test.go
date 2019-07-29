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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
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

func TestReadingNamespaces(t *testing.T) {
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

func TestReadingControllers(t *testing.T) {
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

func TesReadingtAgents(t *testing.T) {
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
	equal = equal && (lhs.KeyFile == rhs.KeyFile)
	equal = equal && (lhs.KubeConfig == rhs.KubeConfig)
	equal = equal && (lhs.KubeControllerIP == rhs.KubeControllerIP)
	equal = equal && (lhs.Name == rhs.Name)
	equal = equal && (lhs.User == lhs.User)
	for key := range lhs.Images {
		equal = equal && (lhs.Images[key] == rhs.Images[key])
	}
	equal = equal && (lhs.IofogUser.Email == lhs.IofogUser.Email)
	equal = equal && (lhs.IofogUser.Name == lhs.IofogUser.Name)
	equal = equal && (lhs.IofogUser.Password == lhs.IofogUser.Password)
	equal = equal && (lhs.IofogUser.Surname == lhs.IofogUser.Surname)

	return equal
}
func TestWritingController(t *testing.T) {
	ctrl := Controller{
		Created:          "Now",
		Endpoint:         "localhost:" + iofog.ControllerPortString,
		Host:             "localhost",
		KeyFile:          "~/.key/file",
		KubeConfig:       "~/.kube/config",
		KubeControllerIP: "123.12.123.13",
		Name:             "Hubert",
		User:             "Kubert",
		Images: map[string]string{
			"controller": "iofog/controller",
			"agent":      "iofog/agent",
			"connector":  "iofog/connector",
			"operator":   "iofog/operator",
		},
		IofogUser: IofogUser{
			Email:    "user@domain.com",
			Name:     "Tubert",
			Password: "NotACockroach",
			Surname:  "Blubert",
		},
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
	equal = equal && (lhs.Image == rhs.Image)
	equal = equal && (lhs.KeyFile == rhs.KeyFile)
	equal = equal && (lhs.Name == rhs.Name)
	equal = equal && (lhs.Port == rhs.Port)
	equal = equal && (lhs.UUID == rhs.UUID)
	equal = equal && (lhs.User == rhs.User)

	return equal
}

func TestWritingAgent(t *testing.T) {
	agent := Agent{
		Created: "Now",
		Host:    "localhost",
		KeyFile: "~/.key/file",
		Name:    "Hubert",
		User:    "Kubert",
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
