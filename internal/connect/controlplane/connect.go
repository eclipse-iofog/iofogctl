package connectcontrolplane

import (
	"context"

	"github.com/eclipse-iofog/iofogctl/internal/auth"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Connect(ctrlPlane rsc.ControlPlane, endpoint, namespace, caFile string, ns *rsc.Namespace) error {
	user := ctrlPlane.GetUser()
	util.SpinHandlePrompt()
	agents, err := auth.ConnectLogin(context.Background(), namespace, endpoint, caFile, user.Email, user.GetRawPassword())
	if err != nil {
		return err
	}
	util.SpinHandlePromptComplete()

	for idx := range agents {
		agent := &agents[idx]
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
