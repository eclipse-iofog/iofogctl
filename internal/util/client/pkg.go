package client

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

var pkg struct {
	// Caches of resources commonly required by iofogctl commands
	clientCache map[string]*client.Client
	agentCache  map[string][]client.AgentInfo
	// Channels for requesting and receiving cached resources
	clientCacheRequestChan chan *clientCacheRequest
	agentCacheRequestChan  chan *agentCacheRequest
	agentSyncRequestChan   chan *agentSyncRequest
}

func init() {
	pkg.clientCache = make(map[string]*client.Client)
	pkg.agentCache = make(map[string][]client.AgentInfo)

	pkg.clientCacheRequestChan = make(chan *clientCacheRequest)
	pkg.agentCacheRequestChan = make(chan *agentCacheRequest)
	pkg.agentSyncRequestChan = make(chan *agentSyncRequest)

	go clientCacheRoutine()
	go agentCacheRoutine()
	go agentSyncRoutine()
}

type clientCacheRequest struct {
	namespace  string
	resultChan chan *clientCacheResult
}

func newClientCacheRequest(namespace string) *clientCacheRequest {
	return &clientCacheRequest{
		namespace:  namespace,
		resultChan: make(chan *clientCacheResult),
	}
}

type clientCacheResult struct {
	err    error
	client *client.Client
}

func (ccr *clientCacheResult) get() (*client.Client, error) {
	return ccr.client, ccr.err
}

type agentCacheRequest struct {
	namespace  string
	invalidate bool
	resultChan chan *agentCacheResult
}

func newAgentCacheRequest(namespace string) *agentCacheRequest {
	return &agentCacheRequest{
		namespace:  namespace,
		resultChan: make(chan *agentCacheResult),
	}
}

func newAgentCacheInvalidateRequest(namespace string) *agentCacheRequest {
	return &agentCacheRequest{
		namespace:  namespace,
		invalidate: true,
		resultChan: make(chan *agentCacheResult),
	}
}

type agentCacheResult struct {
	err    error
	agents []client.AgentInfo
}

func (acr *agentCacheResult) get() ([]client.AgentInfo, error) {
	return acr.agents, acr.err
}

type agentSyncRequest struct {
	namespace  string
	resultChan chan error
}

func newAgentSyncRequest(namespace string) *agentSyncRequest {
	return &agentSyncRequest{
		namespace:  namespace,
		resultChan: make(chan error),
	}
}
