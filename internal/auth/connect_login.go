package auth

import (
	"context"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

// ConnectLogin performs trust-aware iofogUser login and returns controller agents.
func ConnectLogin(ctx context.Context, namespace, endpoint, caFile, email, password string) ([]client.AgentInfo, error) {
	clt, err := newControllerAuthClient(ctx, namespace, endpoint, caFile)
	if err != nil {
		return nil, err
	}
	if err := clt.bootstrapLogin(email, password); err != nil {
		return nil, err
	}
	return clt.listAgents()
}
