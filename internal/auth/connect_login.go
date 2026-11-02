package auth

import (
	"context"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
)

// ConnectSession is the iofog user session established while attaching to a control plane.
type ConnectSession struct {
	Agents       []client.AgentInfo
	AccessToken  string
	RefreshToken string
}

// ConnectLogin performs trust-aware iofogUser login and returns controller agents.
// An empty password or TOTP is prompted by the SDK, matching other user logins.
func ConnectLogin(ctx context.Context, namespace, endpoint, email, password string) (ConnectSession, error) {
	opt, err := trust.SDKOptions(ctx, namespace, endpoint)
	if err != nil {
		return ConnectSession{}, err
	}
	clt, err := client.NewAndLogin(opt, email, password)
	if err != nil {
		return ConnectSession{}, err
	}
	agents, err := clt.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return ConnectSession{}, err
	}
	return ConnectSession{
		Agents:       agents.Agents,
		AccessToken:  clt.GetAccessToken(),
		RefreshToken: clt.GetRefreshToken(),
	}, nil
}
