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

package deletecontroller

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
	ctrl := config.Controller{
		Name: "test_local",
		SSH: config.SSH{
			Host: "localhost",
		},
	}
	if err := config.AddController(ns, ctrl); err != nil {
		t.Errorf("Error when testing local and creating Controller in default namespace: %s", err.Error())
	}
	if _, err := NewExecutor(ns, ctrl.Name); err != nil {
		t.Errorf("Error when testing local and using existing namespace default: %s", err.Error())
	}
}

func TestRemote(t *testing.T) {
	ns := "default"
	ctrl := config.Controller{
		Name: "test_remote",
		Kube: config.Kube{
			Config: "~/.kube/config",
		},
	}
	if err := config.AddController(ns, ctrl); err != nil {
		t.Errorf("Error when testing remote creating Controller in default namespace: %s", err.Error())
	}
	if _, err := NewExecutor(ns, ctrl.Name); err != nil {
		t.Errorf("Error when testing remote and using existing namespace default: %s", err.Error())
	}
}

func TestNonExistentController(t *testing.T) {
	ns := "default"
	ctrlName := "non_existent"
	if _, err := NewExecutor(ns, ctrlName); err == nil {
		t.Error("Expected error with non existent Controller")
	}
}

func TestNonExistentNamespace(t *testing.T) {
	ns := "non_existent"
	ctrlName := "non_existent"
	if _, err := NewExecutor(ns, ctrlName); err == nil {
		t.Error("Expected error with non existent namespace")
	}
}
