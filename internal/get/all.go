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

package get

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
)

type allExecutor struct {
	namespace    string
	showDetached bool
}

func newAllExecutor(namespace string, showDetached bool) *allExecutor {
	exe := &allExecutor{}
	exe.namespace = namespace
	exe.showDetached = showDetached
	return exe
}

func (exe *allExecutor) GetName() string {
	return ""
}

func (exe *allExecutor) Execute() error {
	// Check namespace exists
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	if exe.showDetached {
		printDetached()
		// Print agents
		if err := generateDetachedAgentOutput(); err != nil {
			return err
		}

		return nil
	}
	printNamespace(ns.Name)

	// Print controllers
	if err := generateControllerOutput(exe.namespace, false); err != nil {
		return err
	}

	// Print agents
	if err := generateAgentOutput(exe.namespace, false); err != nil {
		return err
	}

	// Print applications
	appExe := newApplicationExecutor(exe.namespace)
	if err := appExe.init(); err != nil {
		return err
	}
	if err := appExe.generateApplicationOutput(); err != nil {
		return err
	}

	// Print microservices
	msvcExe := newMicroserviceExecutor(exe.namespace)
	if err := msvcExe.init(); err != nil {
		return err
	}
	if err := msvcExe.generateMicroserviceOutput(); err != nil {
		return err
	}

	// Print volumes
	if err := generateVolumeOutput(exe.namespace, false); err != nil {
		return err
	}

	return nil
}

func printDetached() {
	fmt.Printf("DETACHED RESOURCES\n\n")
}
