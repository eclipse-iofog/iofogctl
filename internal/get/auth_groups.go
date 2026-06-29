package get

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/authgroup"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type authGroupsExecutor struct {
	namespace string
}

func newAuthGroupsExecutor(namespace string) *authGroupsExecutor {
	return &authGroupsExecutor{namespace: namespace}
}

func (exe *authGroupsExecutor) GetName() string { return "" }

func (exe *authGroupsExecutor) Execute() error {
	printNamespace(exe.namespace)

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	groups, err := clt.ListAuthGroups()
	if err != nil {
		return authgroup.MapError("", err)
	}

	return print(authgroup.Table(groups))
}

type authGroupExecutor struct {
	namespace string
	name      string
}

func newAuthGroupExecutor(namespace, name string) *authGroupExecutor {
	return &authGroupExecutor{namespace: namespace, name: name}
}

func (exe *authGroupExecutor) GetName() string { return exe.name }

func (exe *authGroupExecutor) Execute() error {
	printNamespace(exe.namespace)

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	group, err := clt.GetAuthGroup(exe.name)
	if err != nil {
		return authgroup.MapError(exe.name, err)
	}

	return print(authgroup.Table([]client.AuthGroupResponse{group}))
}
