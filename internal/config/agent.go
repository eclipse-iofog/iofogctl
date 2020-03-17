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
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

// GetAgents returns all agents within the namespace
func GetAgents(namespace string) ([]Agent, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.Agents, nil
}

// GetAgent returns a single agent within a namespace
func GetAgent(namespace, name string) (agent Agent, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ag := range ns.Agents {
		if ag.Name == name {
			agent = ag
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// Overwrites or creates new agent to the namespace
func UpdateAgent(namespace string, agent Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	// Update existing agent if exists
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == agent.Name {
			mux.Lock()
			ns.Agents[idx] = agent
			mux.Unlock()
			return nil
		}
	}
	// Add new agent
	return AddAgent(namespace, agent)
}

// AddAgent adds a new agent to the namespace
func AddAgent(namespace string, agent Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = GetAgent(namespace, agent.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + agent.Name)
	}

	mux.Lock()
	ns.Agents = append(ns.Agents, agent)
	mux.Unlock()

	return nil
}

// DeleteAgent deletes an agent from a namespace
func DeleteAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == name {
			mux.Lock()
			ns.Agents = append(ns.Agents[:idx], ns.Agents[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}

func GetDetachedAgents() ([]Agent, error) {
	return GetAgents(detachedNamespace)
}

func GetDetachedAgent(name string) (Agent, error) {
	return GetAgent(detachedNamespace, name)
}

func AttachAgent(namespace, name, UUID string) error {
	agent, err := GetDetachedAgent(name)
	if err != nil {
		return err
	}
	if err := DeleteAgent(detachedNamespace, name); err != nil {
		return err
	}
	agent.UUID = UUID
	return UpdateAgent(namespace, agent)
}

func DetachAgent(namespace, name string) error {
	agent, err := GetAgent(namespace, name)
	if err != nil {
		return err
	}
	agent.UUID = ""
	if err := AddAgent(detachedNamespace, agent); err != nil {
		return err
	}
	return DeleteAgent(namespace, name)
}

func RenameDetachedAgent(oldName, newName string) error {
	agent, err := GetDetachedAgent(oldName)
	if err != nil {
		return err
	}
	if err = DeleteDetachedAgent(oldName); err != nil {
		return err
	}
	agent.Name = newName
	return AddAgent(detachedNamespace, agent)
}

func DeleteDetachedAgent(name string) error {
	return DeleteAgent(detachedNamespace, name)
}

func UpdateDetachedAgent(agent Agent) error {
	return UpdateAgent(detachedNamespace, agent)
}
