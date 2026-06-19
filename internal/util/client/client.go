package client

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// clientCacheRoutine handles concurrent requests for a cached Controller client
func clientCacheRoutine() {
	for {
		request := <-pkg.clientCacheRequestChan
		// Invalidate cache
		if request.namespace == "" {
			pkg.clientCache = make(map[string]*client.Client)
			continue
		}
		result := &clientCacheResult{}
		// From cache
		if cachedClient, exists := pkg.clientCache[request.namespace]; exists {
			result.client = cachedClient
			request.resultChan <- result
			continue
		}
		// Create new client
		ioClient, err := newControllerClient(request.namespace)
		// Failure
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Save to cache and return new client
		pkg.clientCache[request.namespace] = ioClient
		result.client = ioClient
		request.resultChan <- result
	}
}

// agentCacheRoutine handles concurrent requests for a cached list of Agents
func agentCacheRoutine() {
	for {
		request := <-pkg.agentCacheRequestChan
		if request.namespace == "" {
			// Invalidate cache
			pkg.agentCache = make(map[string][]client.AgentInfo)
			continue
		}
		result := &agentCacheResult{}
		// From cache
		if cachedAgents, exist := pkg.agentCache[request.namespace]; exist {
			result.agents = cachedAgents
			request.resultChan <- result
			continue
		}
		// Client to get agents
		ioClient, err := NewControllerClient(request.namespace)
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Get agents
		agents, err := getBackendAgents(request.namespace, ioClient)
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Save to cache and return new agents
		pkg.agentCache[request.namespace] = agents
		result.agents = agents
		request.resultChan <- result
	}
}

func agentSyncRoutine() {
	complete := false
	for {
		request := <-pkg.agentSyncRequestChan
		if complete {
			request.resultChan <- nil
			continue
		}
		if err := syncAgentInfo(request.namespace); err != nil {
			request.resultChan <- err
			continue
		}
		complete = true
		request.resultChan <- nil
	}
}

func syncAgentInfo(namespace string) error {
	// Get local cache Agents
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	// Check the Control Plane type
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	if _, ok := controlPlane.(*rsc.LocalControlPlane); ok {
		// Do not update local Agents
		return nil
	}
	// Generate map of config Agents
	agentsMap := make(map[string]*rsc.RemoteAgent)
	var localAgent *rsc.LocalAgent
	for _, baseAgent := range ns.GetAgents() {
		if v, ok := baseAgent.(*rsc.LocalAgent); ok {
			localAgent = v
		} else {
			agentsMap[baseAgent.GetName()] = baseAgent.(*rsc.RemoteAgent)
		}
	}

	// Get backend Agents
	backendAgents, err := GetBackendAgents(namespace)
	if err != nil {
		return err
	}

	// Generate cache types
	agents := make([]rsc.RemoteAgent, len(backendAgents))
	for idx := range backendAgents {
		backendAgent := &backendAgents[idx]
		if localAgent != nil && backendAgent.Name == localAgent.Name {
			localAgent.UUID = backendAgent.UUID
			continue
		}

		agent := rsc.RemoteAgent{
			Name: backendAgent.Name,
			UUID: backendAgent.UUID,
			Host: backendAgent.Host,
		}
		// Update additional info if local cache contains it
		if cachedAgent, exists := agentsMap[backendAgent.Name]; exists {
			agent.Created = cachedAgent.GetCreatedTime()
			agent.SSH = cachedAgent.SSH
		}

		agents[idx] = agent
	}

	// Overwrite the Agents
	ns.DeleteAgents()
	for idx := range agents {
		if err := ns.AddAgent(&agents[idx]); err != nil {
			return err
		}
	}

	if localAgent != nil {
		if err := ns.AddAgent(localAgent); err != nil {
			return err
		}
	}

	return config.Flush()
}

func newControllerClient(namespace string) (*client.Client, error) {

	// Try to get the client from the cache first
	if cachedClient, found := pkg.clientCache[namespace]; found {

		// If a cached client exists, use SessionLogin to refresh the session
		ns, err := config.GetNamespace(namespace)
		if err != nil {
			return nil, err
		}

		controlPlane, err := ns.GetControlPlane()
		if err != nil {
			return nil, err
		}

		endpoint, err := controlPlane.GetEndpoint()
		if err != nil {
			return nil, err
		}

		user := controlPlane.GetUser()

		// Get base URL
		baseURL, err := util.GetBaseURL(endpoint)
		if err != nil {
			return nil, err
		}

		// Use the refresh token from the cached client
		refreshToken := cachedClient.GetRefreshToken()
		user.AccessToken = cachedClient.GetAccessToken()
		user.RefreshToken = cachedClient.GetRefreshToken()
		// controlPlane.UpdateUserTokens(user.AccessToken, user.RefreshToken)
		config.UpdateUser(namespace, user.AccessToken, user.RefreshToken)

		// Use SessionLogin to attempt to refresh the session
		util.SpinHandlePrompt()
		refreshedClient, err := client.SessionLogin(client.Options{BaseURL: baseURL}, refreshToken, user.Email, user.GetRawPassword())
		if err != nil {
			fmt.Println("Error: Failed to refresh session:", err)
			return nil, fmt.Errorf("failed to refresh session: %v", err)
		}
		util.SpinHandlePromptComplete()
		// Update the cached client with the refreshed session
		pkg.clientCache[namespace] = refreshedClient
		config.Flush()
		return refreshedClient, nil
	}

	// If no cached client, proceed with NewAndLogin to create a new client
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return nil, err
	}

	user := controlPlane.GetUser()
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return nil, err
	}

	// Create a new client and login
	util.SpinHandlePrompt()
	newClient, err := client.SessionLogin(client.Options{BaseURL: baseURL}, user.RefreshToken, user.Email, user.GetRawPassword())
	if err != nil {
		return nil, err
	}
	util.SpinHandlePromptComplete()
	user.AccessToken = newClient.GetAccessToken()
	user.RefreshToken = newClient.GetRefreshToken()
	// controlPlane.UpdateUserTokens(user.AccessToken, user.RefreshToken)
	config.UpdateUser(namespace, user.AccessToken, user.RefreshToken)

	// Flush the config and handle errors
	if err := config.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush config: %v", err)
	}

	return newClient, nil
}

func getBackendAgents(namespace string, ioClient *client.Client) ([]client.AgentInfo, error) {
	var agentList client.ListAgentsResponse
	var err error

	// Try the operation
	agentList, err = ioClient.ListAgents(client.ListAgentsRequest{})
	if err == nil {
		pkg.agentCache[namespace] = agentList.Agents
		return agentList.Agents, nil
	}

	// Check if it's an authentication error
	if !isAuthenticationError(err) {
		return nil, err
	}

	// Refresh authentication and retry
	refreshedClient, refreshErr := refreshClientAuthentication(namespace)
	if refreshErr != nil {
		return nil, fmt.Errorf("authentication error occurred and failed to refresh: %v (refresh error: %v)", err, refreshErr)
	}

	// Retry the operation with refreshed client
	agentList, err = refreshedClient.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return nil, err
	}
	pkg.agentCache[namespace] = agentList.Agents
	return agentList.Agents, nil
}

func getRouterAgentNameFromUUID(agentMapByUUID map[string]client.AgentInfo, uuid string) (name string) {
	// if uuid == iofog.VanillaRemoteAgentName {
	// 	return uuid
	// }
	if uuid == iofog.VanillaRouterAgentName {
		return uuid
	}
	agent, found := agentMapByUUID[uuid]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not find Router: %s\n", uuid))
		name = "UNKNOWN ROUTER: " + uuid
	} else {
		name = agent.Name
	}
	return
}

func getNatsAgentNameFromUUID(agentMapByUUID map[string]client.AgentInfo, uuid string) (name string) {

	if uuid == iofog.VanillaNatsAgentName {
		return uuid
	}
	agent, found := agentMapByUUID[uuid]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not find NATS: %s\n", uuid))
		name = "UNKNOWN NATS: " + uuid
	} else {
		name = agent.Name
	}
	return
}

// isAuthenticationError checks if an error is an authentication/authorization error
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Check for common authentication error patterns
	return strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "accessdenied") ||
		strings.Contains(errStr, "Access Denied") ||
		strings.Contains(errStr, "AccessDenied") ||
		strings.Contains(errStr, "Access denied") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "Unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "Forbidden") ||
		strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "Authentication failed") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "token expired") ||
		strings.Contains(errStr, "tokenexpired") ||
		strings.Contains(errStr, "invalid token") ||
		strings.Contains(errStr, "invalidtoken")
}

// refreshClientAuthentication refreshes the client authentication for a namespace
func refreshClientAuthentication(namespace string) (*client.Client, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}

	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}

	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return nil, err
	}

	user := controlPlane.GetUser()
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return nil, err
	}

	// Re-authenticate using SessionLogin
	util.SpinHandlePrompt()
	refreshedClient, err := client.SessionLogin(client.Options{BaseURL: baseURL}, user.RefreshToken, user.Email, user.GetRawPassword())
	if err != nil {
		util.SpinHandlePromptComplete()
		return nil, fmt.Errorf("failed to refresh authentication: %v", err)
	}
	util.SpinHandlePromptComplete()

	// Update tokens in config
	user.AccessToken = refreshedClient.GetAccessToken()
	user.RefreshToken = refreshedClient.GetRefreshToken()
	config.UpdateUser(namespace, user.AccessToken, user.RefreshToken)

	// Update cached client
	pkg.clientCache[namespace] = refreshedClient

	// Flush config
	if err := config.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush config: %v", err)
	}

	return refreshedClient, nil
}

// ExecuteWithAuthRetry executes a function that uses a client, and retries with refreshed auth if auth error occurs
func ExecuteWithAuthRetry(namespace string, operation func(*client.Client) error) error {
	// Get client
	ioClient, err := NewControllerClient(namespace)
	if err != nil {
		return err
	}

	// Try the operation
	err = operation(ioClient)
	if err == nil {
		return nil
	}

	// Check if it's an authentication error
	if !isAuthenticationError(err) {
		return err
	}

	// Refresh authentication and retry
	refreshedClient, refreshErr := refreshClientAuthentication(namespace)
	if refreshErr != nil {
		return fmt.Errorf("authentication error occurred and failed to refresh: %v (refresh error: %v)", err, refreshErr)
	}

	// Retry the operation with refreshed client
	return operation(refreshedClient)
}
