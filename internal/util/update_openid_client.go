package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/eclipse-iofog/iofogctl/internal/resource"
)

// UpdateECNViewerClientRootURL updates the root URL for the ecnviewerclient
// using the controller client secret to obtain an admin token via OAuth2
func UpdateECNViewerClientRootURL(auth resource.Auth, newRootURL string) error {
	// Validate input parameters
	if auth.URL == "" {
		return fmt.Errorf("auth URL is required")
	}
	if auth.Realm == "" {
		return fmt.Errorf("auth realm is required")
	}
	if auth.ControllerClient == "" {
		return fmt.Errorf("controller client ID is required")
	}
	if auth.ControllerSecret == "" {
		return fmt.Errorf("controller client secret is required")
	}
	if auth.ViewerClient == "" {
		return fmt.Errorf("viewer client ID is required")
	}
	if newRootURL == "" {
		return fmt.Errorf("new root URL is required")
	}

	// Configure OAuth2 client credentials
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", auth.URL, auth.Realm)
	config := &clientcredentials.Config{
		ClientID:     auth.ControllerClient,
		ClientSecret: auth.ControllerSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Obtain access token
	ctx := context.Background()
	token, err := config.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain access token: %w", err)
	}

	// Get the client ID (internal Keycloak ID) for the viewer client
	clientID, err := getKeycloakClientID(auth, auth.ViewerClient, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get client ID: %w", err)
	}

	// Update the root URL directly
	err = updateClientRootURL(auth, clientID, newRootURL, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to update client root URL: %w", err)
	}

	return nil
}

// getKeycloakClientID retrieves the internal Keycloak ID for a client by its clientId
func getKeycloakClientID(auth resource.Auth, clientID, adminToken string) (string, error) {
	// Construct admin API URL
	adminURL := fmt.Sprintf("%s/admin/realms/%s/clients", auth.URL, auth.Realm)

	// Create request
	req, err := http.NewRequest("GET", adminURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create get client request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute get client request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get client request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to find the specific client
	var clients []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return "", fmt.Errorf("failed to decode clients response: %w", err)
	}

	// Find the client with matching clientID
	for _, client := range clients {
		if clientIDValue, ok := client["clientId"].(string); ok && clientIDValue == clientID {
			if id, ok := client["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("client with ID '%s' not found", clientID)
}

// updateClientRootURL updates the root URL for a specific client using its internal ID
func updateClientRootURL(auth resource.Auth, clientID, newRootURL, adminToken string) error {
	// Construct admin API URL
	adminURL := fmt.Sprintf("%s/admin/realms/%s/clients/%s", auth.URL, auth.Realm, clientID)

	// Create the update payload with only the root URL
	updatePayload := map[string]interface{}{
		"rootUrl": newRootURL,
	}

	// Marshal to JSON
	payloadJSON, err := json.Marshal(updatePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("PUT", adminURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create update client request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute update client request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update client request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
