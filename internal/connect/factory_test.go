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
		t.Errorf("Error when using existing namespace %s: %s", opt.Namespace, err.Error())
	}
}

func TestNonExistentNamespace(t *testing.T) {
	opt := baseOpt
	opt.Namespace = "connect_factory_test"
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error when using existing namespace %s: %s", opt.Namespace, err.Error())
	}
	if _, err := config.GetNamespace(opt.Namespace); err != nil {
		t.Errorf("Error retrieving namespace: %s", err.Error())
	}
}

func TestNonEmptyNamespace(t *testing.T) {
	opt := baseOpt
	if err := config.AddController(opt.Namespace, config.Controller{}); err != nil {
		t.Errorf("Error adding Controller to config")
	}
	if _, err := NewExecutor(&opt); err == nil {
		t.Errorf("Expected error when using non-empty namespace")
	}
}

func TestOverwiteNamespace(t *testing.T) {
	opt := baseOpt
	if err := config.AddAgent(opt.Namespace, config.Agent{}); err != nil {
		t.Errorf("Error adding Agent to config")
	}
	opt.OverwriteNamespace = true
	if _, err := NewExecutor(&opt); err != nil {
		t.Errorf("Error overwriting existing namespace %s: %s", opt.Namespace, err.Error())
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
