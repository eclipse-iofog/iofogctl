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

package logs

import (
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type kubernetesControllerExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
	name         string
}

func newKubernetesControllerExecutor(controlPlane *rsc.KubernetesControlPlane, namespace, name string) *kubernetesControllerExecutor {
	return &kubernetesControllerExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *kubernetesControllerExecutor) GetName() string {
	return exe.name
}

func (exe *kubernetesControllerExecutor) Execute() error {
	if err := exe.controlPlane.ValidateKubeConfig(); err != nil {
		return err
	}
	out, err := util.Exec("KUBECONFIG="+exe.controlPlane.KubeConfig, "kubectl", "logs", "-l", "name=controller", "-n", exe.namespace)
	if err != nil {
		return err
	}
	fmt.Print(out.String())

	return nil
}
