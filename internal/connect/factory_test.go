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

package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"testing"
)

var baseOpt = Options{
	Namespace: "default",
	Email:     "user@domain.com",
	Password:  "ui89POKJ324!!",
	Endpoint:  "localhost:" + iofog.ControllerPortString,
}

func TestEmptyNamespace(t *testing.T) {
	opt := baseOpt
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when using existing namespace default: %s", err.Error())
	}
}

func TestKubeFile(t *testing.T) {
	opt := baseOpt
	opt.Endpoint = ""
	opt.KubeFile = "~/.kube/config"
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when using existing namespace default: %s", err.Error())
	}
}

func TestNonExistentNamespace(t *testing.T) {
	opt := baseOpt
	opt.Namespace = "connect_factory_test"
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when using existing namespace default: %s", err.Error())
	}
	if _, err := config.GetNamespace(opt.Namespace); err != nil {
		t.Errorf("Error retrieving namespace: %s", err.Error())
	}
}

func TestNoEmail(t *testing.T) {
	opt := baseOpt
	opt.Email = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error with no email specified")
	}
}

func TestNoPassword(t *testing.T) {
	opt := baseOpt
	opt.Password = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error with no password specified")
	}
}

func TestNoController(t *testing.T) {
	opt := baseOpt
	opt.Endpoint = ""
	opt.KubeFile = ""
	if _, err := NewExecutor(&opt); err == nil {
		t.Error("Expected error with no Controller Kube file or endpoint specified")
	}
}
