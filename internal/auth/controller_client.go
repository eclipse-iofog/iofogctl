package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const controllerRequestTimeout = 30 * time.Second

type controllerAuthClient struct {
	baseURL     *url.URL
	httpClient  *http.Client
	accessToken string
}

func newControllerAuthClient(ctx context.Context, namespace, endpoint, caFile string) (*controllerAuthClient, error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return nil, err
	}
	cfg, err := trust.ResolveConnectTransport(ctx, namespace, endpoint, caFile)
	if err != nil {
		return nil, err
	}
	return newControllerAuthClientWithTransport(baseURL, cfg), nil
}

func newControllerAuthClientWithTransport(baseURL *url.URL, cfg trust.TransportConfig) *controllerAuthClient {
	return &controllerAuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: controllerRequestTimeout,
			Transport: &http.Transport{
				TLSClientConfig: cfg.TLSConfig,
			},
		},
	}
}

func (c *controllerAuthClient) bootstrapLogin(email, password string) error {
	body, err := c.doJSON(http.MethodPost, "/user/login", map[string]string{
		"Content-Type": "application/json",
	}, bootstrapLoginRequest{Email: email, Password: password})
	if err != nil {
		return fmt.Errorf("bootstrap login: %w", err)
	}
	var resp client.LoginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parse login response: %w", err)
	}
	c.accessToken = resp.AccessToken
	return nil
}

func (c *controllerAuthClient) listAuthUsers() ([]client.AuthUserResponse, error) {
	body, err := c.doJSON(http.MethodGet, "/users", authHeaders(c.accessToken), nil)
	if err != nil {
		return nil, err
	}
	var users []client.AuthUserResponse
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("parse auth users: %w", err)
	}
	return users, nil
}

func (c *controllerAuthClient) createAuthUser(req client.AuthUserCreateRequest) (client.AuthUserResponse, error) {
	body, err := c.doJSON(http.MethodPost, "/users", authHeaders(c.accessToken), req)
	if err != nil {
		return client.AuthUserResponse{}, err
	}
	var user client.AuthUserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return client.AuthUserResponse{}, fmt.Errorf("parse auth user response: %w", err)
	}
	return user, nil
}

func (c *controllerAuthClient) resetAuthUserToken(userID string) (client.AuthUserResetTokenResponse, error) {
	body, err := c.doJSON(http.MethodPost, fmt.Sprintf("/users/%s/reset-token", userID), authHeaders(c.accessToken), nil)
	if err != nil {
		return client.AuthUserResetTokenResponse{}, err
	}
	var resp client.AuthUserResetTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return client.AuthUserResetTokenResponse{}, fmt.Errorf("parse reset token response: %w", err)
	}
	return resp, nil
}

func (c *controllerAuthClient) changePassword(req client.ChangePasswordRequest) error {
	headers := map[string]string{"Content-Type": "application/json"}
	if req.ResetToken == "" {
		headers = authHeaders(c.accessToken)
	}
	_, err := c.doJSON(http.MethodPost, "/user/change-password", headers, req)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	return nil
}

func (c *controllerAuthClient) listAgents() ([]client.AgentInfo, error) {
	body, err := c.doJSON(http.MethodGet, "/iofog-list", authHeaders(c.accessToken), nil)
	if err != nil {
		return nil, err
	}
	var response client.ListAgentsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parse agents response: %w", err)
	}
	return response.Agents, nil
}

type bootstrapLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func authHeaders(token string) map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}
}

func (c *controllerAuthClient) doJSON(method, requestPath string, headers map[string]string, body any) ([]byte, error) {
	if headers == nil {
		headers = map[string]string{"Content-Type": "application/json"}
	}
	requestURL := *c.baseURL
	switch parts := splitPathQuery(requestPath); len(parts) {
	case 1:
		requestURL.Path = path.Join(requestURL.Path, parts[0])
	case 2:
		requestURL.Path = path.Join(requestURL.Path, parts[0])
		requestURL.RawQuery = parts[1]
	default:
		return nil, fmt.Errorf("invalid request path %q", requestPath)
	}

	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, requestURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s returned %d: %s", method, requestPath, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func splitPathQuery(requestPath string) []string {
	for i := 0; i < len(requestPath); i++ {
		if requestPath[i] == '?' {
			return []string{requestPath[:i], requestPath[i+1:]}
		}
	}
	return []string{requestPath}
}
