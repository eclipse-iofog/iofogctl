package connectcontrolplane

import (
	"context"

	"github.com/eclipse-iofog/iofogctl/internal/auth"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// PrepareTrust validates --ca / --ca-b64 flags or spec.ca from YAML and persists trust for the namespace.
func PrepareTrust(namespace string, cp rsc.ControlPlane, caFile, caB64 string) error {
	normalized, err := trust.NormalizeTrustCA(caFile, caB64)
	if err != nil {
		return err
	}
	if normalized != "" {
		if err := rsc.SetTrustCA(cp, normalized); err != nil {
			return err
		}
		return trust.StoreCA(namespace, normalized)
	}
	if ca := rsc.GetTrustCA(cp); ca != "" {
		return trust.StoreCA(namespace, ca)
	}
	return nil
}

func Connect(ctrlPlane rsc.ControlPlane, endpoint, namespace string, ns *rsc.Namespace) error {
	user := ctrlPlane.GetUser()
	util.SpinHandlePrompt()
	session, err := auth.ConnectLogin(context.Background(), namespace, endpoint, user.Email, user.GetRawPassword())
	util.SpinHandlePromptComplete()
	if err != nil {
		return err
	}
	ctrlPlane.UpdateUserTokens(session.AccessToken, session.RefreshToken)

	for idx := range session.Agents {
		agent := &session.Agents[idx]
		agentConfig := rsc.RemoteAgent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err := ns.AddAgent(&agentConfig); err != nil {
			return err
		}
	}
	return nil
}
