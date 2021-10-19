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

type Agent interface {
	GetName() string
	GetUUID() string
	GetHost() string
	GetCreatedTime() string
	GetConfig() *AgentConfiguration
	GetControllerEndpoint() string
	SetName(string)
	SetUUID(string)
	SetHost(string)
	SetCreatedTime(string)
	SetConfig(*AgentConfiguration)
	Sanitize() error
	Clone() Agent
}
