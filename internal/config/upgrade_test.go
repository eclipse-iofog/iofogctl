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
	"testing"

	v1 "github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

var configData = []byte(`kind: IofogctlConfig
apiVersion: iofogctl/v1
spec:
  defaultNamespace: default
`)

const configDir = "/tmp/"
const namespaceDir = "/tmp/namespaces"
const configFile = "/tmp/config.yaml"

func setup() {
	os.RemoveAll(namespaceDir)
	err := ioutil.WriteFile(configFile, configData, 0644)
	if err != nil {
		panic(err)
	}

	v1.Init(configFile)
}
func teardown() {
	os.RemoveAll(namespaceDir)
	os.RemoveAll(configFile)
}

func TestKubernetesNamespace(t *testing.T) {
	setup()

	nsName := "k8s"
	nsV1 := v1.Namespace{
		ControlPlane: v1.ControlPlane{
			IofogUser: v1.IofogUser{
				Email:    "user@domain.com",
				Password: "password",
			},
			Controllers: []v1.Controller{
				{
					Kube: v1.Kube{
						Config: "~/.kube/config",
						Images: v1.KubeImages{
							Operator: "iofog/operator:latest",
							Kubelet:  "iofog/kubelet:latest",
						},
					},
				},
			},
		},
		Agents: []v1.Agent{
			{
				Name: "agent-1",
				Host: "123.123.123.123",
				SSH: v1.SSH{
					Port:    22,
					User:    "serge",
					KeyFile: "~/.ssh/id_rsa",
				},
			},
		},
	}

	run(nsName, nsV1, t)

	teardown()
}

func TestRemoteNamespace(t *testing.T) {
	setup()

	nsName := "remote"
	nsV1 := v1.Namespace{
		ControlPlane: v1.ControlPlane{
			IofogUser: v1.IofogUser{
				Email:    "user@domain.com",
				Password: "password",
			},
			Controllers: []v1.Controller{
				{
					Name: "controller-1",
					Host: "123.123.123.123",
					SSH: v1.SSH{
						Port:    22,
						User:    "serge",
						KeyFile: "~/.ssh/id_rsa",
					},
				},
			},
		},
		Agents: []v1.Agent{
			{
				Name: "agent-1",
				Host: "123.123.123.123",
				SSH: v1.SSH{
					Port:    22,
					User:    "serge",
					KeyFile: "~/.ssh/id_rsa",
				},
			},
		},
	}

	run(nsName, nsV1, t)

	teardown()
}

func TestLocalNamespace(t *testing.T) {
	setup()

	nsName := "local"
	nsV1 := v1.Namespace{
		ControlPlane: v1.ControlPlane{
			IofogUser: v1.IofogUser{
				Email:    "user@domain.com",
				Password: "password",
			},
			Controllers: []v1.Controller{
				{
					Name: "controller-1",
					Host: "localhost",
				},
			},
		},
		Agents: []v1.Agent{
			{
				Name: "agent-1",
				Host: "localhost",
			},
		},
	}

	run(nsName, nsV1, t)

	teardown()
}

func run(nsName string, nsV1 v1.Namespace, t *testing.T) {
	if err := v1.AddNamespace(nsName, util.NowUTC()); err != nil {
		t.Error("Failed to add namespace to v1")
	}
	if err := v1.UpdateControlPlane(nsName, nsV1.ControlPlane); err != nil {
		t.Error("Failed to update Control Plane v1")
	}
	for _, agent := range nsV1.Agents {
		if err := v1.AddAgent(nsName, agent); err != nil {
			t.Errorf("Failed to add Agent v1: %s", err.Error())
		}
	}
	if err := v1.Flush(); err != nil {
		t.Error("Failed to flush v1")
	}

	// V2 stuff
	Init(configDir)
	ns, err := GetNamespace(nsName)
	if err != nil {
		t.Errorf("Failed to get Namespace v2 %s", err.Error())
	}
	cp, err := ns.GetControlPlane()
	if err != nil {
		t.Errorf("Failed to get Control Plane v2: %s", err.Error())
	}
	if len(cp.GetControllers()) == 0 {
		t.Errorf("Failed to find Controllers v2, count: %d", len(cp.GetControllers()))
	}
	if len(ns.GetAgents()) == 0 {
		t.Errorf("Failed to find Agents v2, count: %d", len(ns.GetAgents()))
	}
}
