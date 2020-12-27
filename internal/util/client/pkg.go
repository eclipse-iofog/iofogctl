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

package client

import (
	"sync"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
)

var pkg struct {
	// Caches of resources commonly required by iofogctl commands
	clientCache map[string]*client.Client
	agentCache  map[string][]client.AgentInfo
	// Channels for requesting and receiving cached resources
	clientReqChan chan string
	clientChan    chan clientCacheResult
	agentReqChan  chan string
	agentChan     chan agentCacheResult
	// Once off for sync agent info
	once sync.Once
}

func init() {
	pkg.clientCache = make(map[string]*client.Client)
	pkg.agentCache = make(map[string][]client.AgentInfo)
	pkg.clientReqChan = make(chan string)
	pkg.clientChan = make(chan clientCacheResult)
	pkg.agentReqChan = make(chan string)
	pkg.agentChan = make(chan agentCacheResult)
	go clientCacheRoutine()
	go agentCacheRoutine()
}

type clientCacheResult struct {
	err    error
	client *client.Client
}

func (ccr *clientCacheResult) get() (*client.Client, error) {
	return ccr.client, ccr.err
}

type agentCacheResult struct {
	err    error
	agents []client.AgentInfo
}

func (acr *agentCacheResult) get() ([]client.AgentInfo, error) {
	return acr.agents, acr.err
}
