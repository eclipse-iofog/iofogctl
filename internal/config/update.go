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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

func UpdateNamespace(newNamespace rsc.Namespace) {
	mux.Lock()
	defer mux.Unlock()
	namespaces[newNamespace.Name] = &newNamespace
}

func UpdateControlPlane(namespace string, controlPlane rsc.ControlPlane) {
	mux.Lock()
	defer mux.Unlock()
	namespaces[namespace].SetControlPlane(controlPlane)
}
