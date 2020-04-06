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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

func GetAgent(namespace, name string) (rsc.Agent, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.GetAgent(name)
}

func GetAgents(namespace string) ([]rsc.Agent, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.GetAgents(), nil
}

func AddAgent(namespace string, agent rsc.Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	return ns.AddAgent(agent)
}

func UpdateAgent(namespace string, agent rsc.Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	ns.UpdateAgent(agent)
	return nil
}

func DeleteAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	return ns.DeleteAgent(name)
}

func GetDetachedAgents() []rsc.Agent {
	ns, _ := getNamespace(detachedNamespace)
	return ns.GetAgents()
}

func GetDetachedAgent(name string) (rsc.Agent, error) {
	return GetAgent(detachedNamespace, name)
}

func AttachAgent(namespace, name, UUID string) error {
	detachedAgent, err := GetAgent(detachedNamespace, name)
	if err != nil {
		return err
	}
	if err := DeleteAgent(detachedNamespace, name); err != nil {
		return err
	}
	detachedAgent.SetUUID(UUID)
	return AddAgent(namespace, detachedAgent)
}

func DetachAgent(namespace, name string) error {
	agent, err := GetAgent(namespace, name)
	if err != nil {
		return err
	}
	agent.SetUUID("")
	if err := AddAgent(detachedNamespace, agent); err != nil {
		return err
	}
	return DeleteAgent(namespace, name)
}

func RenameDetachedAgent(oldName, newName string) error {
	detachedAgent, err := GetAgent(detachedNamespace, oldName)
	if err != nil {
		return err
	}
	if err = DeleteAgent(detachedNamespace, oldName); err != nil {
		return err
	}
	detachedAgent.SetName(newName)
	return AddAgent(detachedNamespace, detachedAgent)
}

func DeleteDetachedAgent(name string) error {
	return DeleteAgent(detachedNamespace, name)
}

func UpdateDetachedAgent(agent rsc.Agent) {
	UpdateAgent(detachedNamespace, agent)
}
