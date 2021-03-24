/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package resource

import (
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"testing"
)

const (
	email    = "user@domain.com"
	password = "as901yh3rinsd"
)

func TestKubernetesControlPlane(t *testing.T) {
	cp := KubernetesControlPlane{
		Endpoint:   "123.123.123.123",
		KubeConfig: "~/.kube/config",
		IofogUser:  IofogUser{Email: "user@domain.com", Password: "password"},
	}
	if endpoint, err := cp.GetEndpoint(); err != nil || endpoint != "123.123.123.123" {
		t.Error("Wrong endpoint")
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if err := cp.AddController(&KubernetesController{
		PodName:  "pod1",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}
	if err := cp.AddController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}

	if err := cp.AddController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err == nil {
		t.Error("Should have failed adding duplicate Controller")
	}

	if err := cp.UpdateController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}
	for _, ctrl := range cp.GetControllers() {
		if ctrl.GetCreatedTime() != "now" || ctrl.GetEndpoint() != "123.123.123.123" {
			t.Error("Controller details are wrong")
		}
	}
	if err := cp.DeleteController("pod2"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}
	if err := cp.DeleteController("pod2"); err == nil {
		t.Error("Deleted non existent Controller")
	}
	if err := cp.DeleteController("pod1"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
}

func TestRemoteControlPlane(t *testing.T) {
	cp := RemoteControlPlane{
		IofogUser: IofogUser{Email: "user@domain.com", Password: "password"},
	}
	if err := cp.AddController(&RemoteController{
		Name:     "ctrl1",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}
	if endpoint, err := cp.GetEndpoint(); err != nil || endpoint != "123.123.123.123" {
		t.Error("Wrong endpoint")
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if err := cp.AddController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}

	if err := cp.AddController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err == nil {
		t.Error("Should have failed adding duplicate Controller")
	}

	if err := cp.UpdateController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}
	for _, ctrl := range cp.GetControllers() {
		if ctrl.GetCreatedTime() != "now" || ctrl.GetEndpoint() != "123.123.123.123" {
			t.Error("Controller details are wrong")
		}
	}
	if err := cp.DeleteController("ctrl2"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}
	if err := cp.DeleteController("ctrl2"); err == nil {
		t.Error("Deleted non existent Controller")
	}
	if err := cp.DeleteController("ctrl1"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
}

func TestLocalControlPlane(t *testing.T) {
	cp := LocalControlPlane{
		IofogUser: IofogUser{Email: "user@domain.com", Password: "password"},
	}
	if err := cp.AddController(&LocalController{
		Name:    "ctrl1",
		Created: "now",
	}); err != nil {
		t.Error(err)
	}
	cp.Sanitize()

	if endpoint, err := cp.GetEndpoint(); err != nil || !util.IsLocalHost(endpoint) {
		t.Errorf("Wrong endpoint: %s", endpoint)
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}

	if err := cp.DeleteController(""); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
	if ctrl, err := cp.GetController(""); err == nil || ctrl != nil {
		t.Error("Should have returned error when getting Local Controller")
	}
}
