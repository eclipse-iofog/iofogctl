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

package deployvolume

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type remoteExecutor struct {
	volume rsc.Volume
	ns     *rsc.Namespace
	agents []*rsc.RemoteAgent
}

func (exe *remoteExecutor) GetName() string {
	return "deploying Volume " + exe.volume.Name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart("Pushing volumes to Agents")
	// Transfer files
	nbAgents := len(exe.agents)
	ch := make(chan error, nbAgents)
	for idx := range exe.agents {
		go exe.execute(idx, ch)
	}
	for idx := 0; idx < nbAgents; idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}
	// Update config
	exe.ns.UpdateVolume(&exe.volume)
	return config.Flush()
}

func (exe *remoteExecutor) execute(agentIdx int, ch chan error) {
	agent := exe.agents[agentIdx]

	// Connect
	ssh, err := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err != nil {
		msg := "failed to initialize SSH client %s.\n%s"
		ch <- fmt.Errorf(msg, agent.Name, err.Error())
		return
	}
	if err := ssh.Connect(); err != nil {
		msg := "failed to Connect to Agent %s.\n%s"
		ch <- fmt.Errorf(msg, agent.Name, err.Error())
		return
	}
	defer util.Log(ssh.Disconnect)

	// Create dest dir
	if err := ssh.CreateFolder(exe.volume.Destination); err != nil {
		msg := "failed to create base directory %s on Agent %s.\n%s"
		ch <- fmt.Errorf(msg, exe.volume.Destination, agent.Name, err.Error())
		return
	}
	// Create tmp dir
	tmp := "/tmp/iofogctlssh" + exe.volume.Destination
	if err := ssh.CreateFolder(tmp); err != nil {
		msg := "failed to create base directory %s on Agent %s.\n%s"
		ch <- fmt.Errorf(msg, exe.volume.Destination, agent.Name, err.Error())
		return
	}
	// Copy volume to tmp
	if err := ssh.CopyFolderTo(exe.volume.Source, tmp, exe.volume.Permissions, true); err != nil {
		msg := "failed to copy volume to Agent %s.\n%s"
		ch <- fmt.Errorf(msg, agent.Name, err.Error())
		return
	}
	// Move volume from tmp to dest
	ifStr := fmt.Sprintf(`[ -z "$(ls -A %s)" ]`, tmp)
	mkdirStr := fmt.Sprintf(`mkdir -p %s`, exe.volume.Destination)
	cpStr := fmt.Sprintf(`sudo -S cp -pR %s/* %s`, tmp, exe.volume.Destination)
	if _, err := ssh.Run(fmt.Sprintf("%s && %s || %s", ifStr, mkdirStr, cpStr)); err != nil {
		msg := "failed to move volume from %s to %s on Agent %s.\n%s"
		ch <- fmt.Errorf(msg, tmp, exe.volume.Destination, agent.Name, err.Error())
		return
	}
	// Remove tmp
	if _, err := ssh.Run(fmt.Sprintf("rm -rf %s", tmp)); err != nil {
		msg := "failed clearing tmp volume data %s from Agent %s.\n%s"
		ch <- fmt.Errorf(msg, tmp, agent.Name, err.Error())
		return
	}

	ch <- nil
}
