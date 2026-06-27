package exec

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type agentExecutor struct {
	namespace  string
	name       string
	debugImage *string
}

func newAgentExecutor(namespace, name string, debugImage *string) *agentExecutor {
	return &agentExecutor{
		namespace:  namespace,
		name:       name,
		debugImage: debugImage,
	}
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	label := "Agent " + exe.name
	return runExecSession(exe.namespace, label, func(clt *client.Client) (*client.ExecSession, error) {
		msvc, err := ensureDebugExecReady(clt, exe.name, exe.debugImage)
		if err != nil {
			return nil, err
		}
		return dialExecForMicroservice(clt, msvc, true)
	})
}
