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

package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"testing"
)

var baseOpt = Options{
	Host:      "123.123.123.123",
	KeyFile:   "~/.ssh/id_rsa",
	User:      "username",
	Namespace: "default",
}

func TestCreateNamespaceAndController(t *testing.T) {
	if err := config.AddNamespace(baseOpt.Namespace, ""); err != nil {
		t.Errorf("Error when creating default namespace: %s", err.Error())
	}
	ctrl := config.Controller{
		Name: "First",
	}
	if err := config.AddController(baseOpt.Namespace, ctrl); err != nil {
		t.Errorf("Error when creating a Controller in default namespace: %s", err.Error())
	}
}

func TestRemote(t *testing.T) {
	opt := baseOpt
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when creating remote executor: %s", err.Error())
	}
}

func TestLocal(t *testing.T) {
	opt := baseOpt
	opt.Host = "localhost"
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when creating local executor: %s", err.Error())
	}
}

func TestNonExistentNamespace(t *testing.T) {
	opt := baseOpt
	opt.Namespace = "non_existent"
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error when non-existent namespace is specified")
	}
}

func TestNoHost(t *testing.T) {
	opt := baseOpt
	opt.Host = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error when no host is specified")
	}
}

func TestNoUser(t *testing.T) {
	opt := baseOpt
	opt.User = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error when no user is specified")
	}
}

func TestNoKeyFile(t *testing.T) {
	opt := baseOpt
	opt.KeyFile = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error when no key file is specified")
	}
}

func TestNoControllerInNamespace(t *testing.T) {
	opt := baseOpt
	opt.Namespace = "not_default"
	if err := config.AddNamespace(opt.Namespace, ""); err != nil {
		t.Errorf("Error when creating new namespace: %s", err.Error())
	}
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error in namespace with no Controller")
	}
}

func TestMultipleControllersInNamespace(t *testing.T) {
	ctrl := config.Controller{
		Name: "Multiple",
	}
	if err := config.AddController(baseOpt.Namespace, ctrl); err != nil {
		t.Errorf("Error when creating a Controller in default namespace: %s", err.Error())
	}
	if _, err := NewExecutor(&baseOpt); err == nil {
		t.Error("Expected error in namespace with multiple Controllers")
	}
}
