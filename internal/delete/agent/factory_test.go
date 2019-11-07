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

package deleteagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"testing"
)

func TestCreateNamespace(t *testing.T) {
	if err := config.AddNamespace("default", ""); err != nil {
		t.Errorf("Error when creating default namespace: %s", err.Error())
	}
}

func TestLocal(t *testing.T) {
	ns := "default"
	agent := config.Agent{
		Name: "test_local",
		Host: "localhost",
	}
	if err := config.AddAgent(ns, agent); err != nil {
		t.Errorf("Error when testing local and creating Agent in default namespace: %s", err.Error())
	}
	if _, err := NewExecutor(ns, agent.Name); err != nil {
		t.Errorf("Error when testing local and using existing namespace default: %s", err.Error())
	}
}

func TestRemote(t *testing.T) {
	ns := "default"
	agent := config.Agent{
		Name: "test_remote",
		Host: "123.123.123.123",
		SSH: config.SSH{
			User:    "serge",
			KeyFile: "~/.ssh/id_rsa",
			Port:    22,
		},
	}
	if err := config.AddAgent(ns, agent); err != nil {
		t.Errorf("Error when testing remote creating Agent in default namespace: %s", err.Error())
	}
	if _, err := NewExecutor(ns, agent.Name); err != nil {
		t.Errorf("Error when testing remote and using existing namespace default: %s", err.Error())
	}
}

func TestNonExistentAgent(t *testing.T) {
	ns := "default"
	agentName := "non_existent"
	if _, err := NewExecutor(ns, agentName); err == nil {
		t.Error("Expected error with non existent Agent")
	}
}

func TestNonExistentNamespace(t *testing.T) {
	ns := "non_existent"
	agentName := "non_existent"
	if _, err := NewExecutor(ns, agentName); err == nil {
		t.Error("Expected error with non existent namespace")
	}
}
