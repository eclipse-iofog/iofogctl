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

package util

import (
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
)

func IsSystemAgent(agentConfig *rsc.AgentConfiguration) bool {
	return agentConfig != nil && agentConfig.IsSystem != nil && *agentConfig.IsSystem
}

func MakeIntPtr(value int) *int {
	return &value
}

func MakeStrPtr(value string) *string {
	return &value
}

func MakeBoolPtr(value bool) *bool {
	return &value
}
