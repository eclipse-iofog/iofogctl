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

package install

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func copyDir(src, dst string) (err error) {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}
	for _, file := range files {
		if err = copyFile(path.Join(src, file.Name()), path.Join(dst, file.Name())); err != nil {
			return
		}
	}
	return
}

func copyFile(src, dst string) (err error) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return
	}
	return
}

type testState struct {
	user      string
	host      string
	port      int
	keyFile   string
	agentName string
	agentUUID string
	dir       string
	srcDir    string
	procs     AgentProcedures
}

var state = testState{
	user:      "serge",
	host:      "localhost",
	port:      51121,
	keyFile:   "~/.ssh/id_rsa",
	agentName: "albert",
	agentUUID: "ashdifafhsdiofd",
	srcDir:    "../../../assets/agent",
	dir:       "/tmp/iofogctl-test-go",
}

func runTest(t *testing.T, state testState) {
	agent := RemoteAgent{}

	if err := agent.CustomizeProcedures(state.dir, &state.procs); err != nil {
		t.Fatalf("Failed to customize procedures: %s", err.Error())
	}

	expectFiles, err := ioutil.ReadDir(state.srcDir)
	if err != nil {
		t.Fatalf("Failed to count files in src script dir: %s", err.Error())
	}
	expect := len(expectFiles)
	if len(agent.procs.scriptNames) != expect {
		t.Fatalf("Expected %d scripts names, found %d %v", expect, len(agent.procs.scriptNames), agent.procs.scriptNames)
		if len(agent.procs.scriptContents) != len(agent.procs.scriptNames) {
			t.Fatalf("Expected %d scripts contents, found %d", len(agent.procs.scriptNames), len(agent.procs.scriptContents))
		}
		for idx, filename := range agent.procs.scriptNames {
			fileBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Fatalf("Failed to read script %s: %s", filename, err.Error())
			}
			if string(fileBytes) != agent.procs.scriptContents[idx] {
				t.Fatalf("Script contents for %s are not correct", filename)
			}
		}
	}
}

func generateScripts(t *testing.T, rm []string) {
	os.RemoveAll(state.dir)
	if err := os.MkdirAll(state.dir, os.FileMode(0777)); err != nil {
		t.Fatalf("Failed to create dir: %s", err.Error())
	}
	if err := copyDir(state.srcDir, state.dir); err != nil {
		t.Fatalf("Failed to copy dir: %s", err.Error())
	}
	if err := os.Remove(path.Join(state.dir, pkg.scriptPrereq)); err != nil {
		t.Fatalf("Failed to delete %s: %s", pkg.scriptPrereq, err.Error())
	}
	for _, file := range rm {
		if err := os.Remove(path.Join(state.dir, file)); err != nil {
			t.Fatalf("Failed to delete %s: %s", file, err.Error())
		}
	}
}

func TestCustomProceduresFull(t *testing.T) {
	generateScripts(t, []string{})
	state.procs = AgentProcedures{
		Deps: Entrypoint{
			Name: pkg.scriptInstallDeps,
		},
		Install: Entrypoint{
			Name: pkg.scriptInstallIofog,
			Args: []string{
				"",
				"",
				"",
			},
		},
		Uninstall: Entrypoint{
			Name: pkg.scriptUninstallIofog,
		},
	}
	runTest(t, state)
}

func TestCustomProceduresPartial(t *testing.T) {
	generateScripts(t, []string{pkg.scriptInstallIofog})
	state.procs = AgentProcedures{
		Deps: Entrypoint{
			Name: pkg.scriptInstallDeps,
		},
		Uninstall: Entrypoint{
			Name: pkg.scriptUninstallIofog,
		},
	}
	runTest(t, state)

	//	generateScripts(t, []string{pkg.scriptInstallIofog, pkg.scriptInstallDeps, pkg.scriptInstallDocker, pkg.scriptInstallJava})
	//	state.procs = AgentProcedures{
	//		Uninstall: Entrypoint{
	//			Name: pkg.scriptUninstallIofog,
	//		},
	//	}
	//	runTest(t, state)
	//
	//	generateScripts(t, []string{pkg.scriptUninstallIofog, pkg.scriptInstallIofog, pkg.scriptInstallDeps, pkg.scriptInstallDocker, pkg.scriptInstallJava})
	//	state.procs = AgentProcedures{}
	//	runTest(t, state)
	//
	//	generateScripts(t, []string{pkg.scriptUninstallIofog, pkg.scriptInstallDeps, pkg.scriptInstallDocker, pkg.scriptInstallJava})
	//	state.procs = AgentProcedures{
	//		Install: Entrypoint{
	//			Name: pkg.scriptInstallIofog,
	//			Args: []string{
	//				"",
	//				"",
	//				"",
	//			},
	//		},
	//	}
	//	runTest(t, state)
}
