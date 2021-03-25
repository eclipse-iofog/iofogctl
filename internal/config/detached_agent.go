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

func findDetachedAgentVersion(agentName string) (version string, err error) {
	for _, vers := range pkg.supportedVersions {
		var ns *rsc.Namespace
		if ns, err = getNamespace(pkg.detachedNamespace, vers); err != nil {
			return
		}
		if _, err = ns.GetAgent(agentName); err == nil {
			version = vers
			return
		}
	}
	return
}

func GetDetachedAgent(name string) (rsc.Agent, error) {
	version, err := findDetachedAgentVersion(name)
	if err != nil {
		return nil, err
	}
	return getDetachedAgent(name, version)
}

func getDetachedAgent(name, version string) (rsc.Agent, error) {
	ns, err := getNamespace(pkg.detachedNamespace, version)
	if err != nil {
		return nil, err
	}
	return ns.GetAgent(name)
}

func GetDetachedAgents() (agents []rsc.Agent, err error) {
	for _, vers := range pkg.supportedVersions {
		var ns *rsc.Namespace
		if ns, err = getNamespace(pkg.detachedNamespace, vers); err == nil {
			agents = ns.GetAgents()
			return
		}
	}
	return
}

func AttachAgent(namespace, name, uuid string) error {
	version, err := findDetachedAgentVersion(name)
	if err != nil {
		return err
	}
	ns, err := getNamespace(namespace, version)
	if err != nil {
		return err
	}
	detachedAgent, err := getDetachedAgent(name, version)
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
	version, err := findDetachedAgentVersion(name)
	if err != nil {
		return err
	}
	ns, err := getNamespace(namespace, version)
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
	version, err := findDetachedAgentVersion(agent.GetName())
	if err != nil {
		return err
	}
	return addDetachedAgent(agent, version)
}

func addDetachedAgent(agent rsc.Agent, version string) error {
	ns, err := getNamespace(pkg.detachedNamespace, version)
	if err != nil {
		return err
	}
	return ns.AddAgent(agent)
}

func RenameDetachedAgent(oldName, newName string) error {
	version, err := findDetachedAgentVersion(oldName)
	if err != nil {
		return err
	}
	detachedAgent, err := getDetachedAgent(oldName, version)
	if err != nil {
		return err
	}
	if err := deleteDetachedAgent(oldName, version); err != nil {
		return err
	}
	detachedAgent.SetName(newName)
	return addDetachedAgent(detachedAgent, version)
}

func DeleteDetachedAgent(name string) error {
	version, err := findDetachedAgentVersion(name)
	if err != nil {
		return err
	}
	return deleteDetachedAgent(name, version)
}

func deleteDetachedAgent(name, version string) error {
	ns, err := getNamespace(pkg.detachedNamespace, version)
	if err != nil {
		return err
	}
	return ns.DeleteAgent(name)
}

func UpdateDetachedAgent(agent rsc.Agent) error {
	version, err := findDetachedAgentVersion(agent.GetName())
	if err != nil {
		return err
	}
	return updateDetachedAgent(agent, version)
}

func updateDetachedAgent(agent rsc.Agent, version string) error {
	ns, err := getNamespace(pkg.detachedNamespace, version)
	if err != nil {
		return err
	}
	return ns.UpdateAgent(agent)
}
