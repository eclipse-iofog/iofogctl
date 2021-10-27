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

package config

import (
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
)

func GetDetachedAgent(name string) (rsc.Agent, error) {
	ns, err := getNamespace(detachedNamespace)
	if err != nil {
		return nil, err
	}
	return ns.GetAgent(name)
}

func GetDetachedAgents() []rsc.Agent {
	ns, _ := getNamespace(detachedNamespace)
	return ns.GetAgents()
}

func AttachAgent(namespace, name, uuid string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	detachedAgent, err := GetDetachedAgent(name)
	if err != nil {
		return err
	}
	if err := DeleteDetachedAgent(name); err != nil {
		return err
	}
	detachedAgent.SetUUID(uuid)
	return ns.AddAgent(detachedAgent)
}

func DetachAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	agent, err := ns.GetAgent(name)
	if err != nil {
		return err
	}
	agent.SetUUID("")
	if err := AddDetachedAgent(agent); err != nil {
		return err
	}
	return ns.DeleteAgent(name)
}

func AddDetachedAgent(agent rsc.Agent) error {
	ns, err := getNamespace(detachedNamespace)
	if err != nil {
		return err
	}
	return ns.AddAgent(agent)
}

func RenameDetachedAgent(oldName, newName string) error {
	detachedAgent, err := GetDetachedAgent(oldName)
	if err != nil {
		return err
	}
	if err := DeleteDetachedAgent(oldName); err != nil {
		return err
	}
	detachedAgent.SetName(newName)
	return AddDetachedAgent(detachedAgent)
}

func DeleteDetachedAgent(name string) error {
	ns, err := getNamespace(detachedNamespace)
	if err != nil {
		return err
	}
	return ns.DeleteAgent(name)
}

func UpdateDetachedAgent(agent rsc.Agent) error {
	ns, err := getNamespace(detachedNamespace)
	if err != nil {
		return err
	}
	return ns.UpdateAgent(agent)
}
