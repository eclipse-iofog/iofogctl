// Package docker wraps the Moby API for legacy local control plane container operations.
// TODO(v3.8.0): remove after local and remote control plane no longer use Go container deploy.
package docker

import (
	"context"
	"fmt"
	"strings"

	mobyclient "github.com/moby/moby/client"
)

// Client wraps the Moby Docker API client.
type Client struct {
	cli *mobyclient.Client
}

// NewWithHost connects to a container engine using a unix/tcp host URL.
// When hostURL is empty, Moby falls back to DOCKER_HOST and default socket discovery.
func NewWithHost(hostURL string) (*Client, error) {
	var opts []mobyclient.Opt
	if strings.TrimSpace(hostURL) != "" {
		opts = append(opts, mobyclient.WithHost(hostURL))
	} else {
		opts = append(opts, mobyclient.FromEnv)
	}
	cli, err := mobyclient.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// Close releases the underlying client.
func (c *Client) Close() error {
	if c == nil || c.cli == nil {
		return nil
	}
	return c.cli.Close()
}

func (c *Client) context() context.Context {
	return context.Background()
}
