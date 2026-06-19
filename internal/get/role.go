package get

import (
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type roleExecutor struct {
	namespace string
}

func newRoleExecutor(namespace string) *roleExecutor {
	a := &roleExecutor{}
	a.namespace = namespace
	return a
}

func (exe *roleExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateRolesOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *roleExecutor) GetName() string {
	return ""
}

func generateRolesOutput(namespace string) error {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	list, err := clt.ListRoles()
	if err != nil {
		return err
	}

	return tabulateRoles(list.Roles)
}

func tabulateRoles(roles []client.RoleInfo) error {
	table := make([][]string, len(roles)+1)
	headers := []string{
		"NAME",
		"KIND",
		"RULES",
	}
	table[0] = append(table[0], headers...)

	for idx, r := range roles {
		rulesCount := strconv.Itoa(len(r.Rules))
		row := []string{
			r.Name,
			r.Kind,
			rulesCount,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return print(table)
}
