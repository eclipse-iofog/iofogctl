package get

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type serviceAccountExecutor struct {
	namespace string
}

func newServiceAccountExecutor(namespace string) *serviceAccountExecutor {
	a := &serviceAccountExecutor{}
	a.namespace = namespace
	return a
}

func (exe *serviceAccountExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateServiceAccountsOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *serviceAccountExecutor) GetName() string {
	return ""
}

func generateServiceAccountsOutput(namespace string) error {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	list, err := clt.ListServiceAccounts("")
	if err != nil {
		return err
	}

	return tabulateServiceAccounts(list.ServiceAccounts)
}

func tabulateServiceAccounts(accounts []client.ServiceAccountInfo) error {
	table := make([][]string, len(accounts)+1)
	headers := []string{
		"APPLICATION",
		"NAME",
		"ROLE",
	}
	table[0] = append(table[0], headers...)

	for idx, sa := range accounts {
		roleName := sa.RoleRef.Name
		appName := sa.ApplicationName
		row := []string{
			appName,
			sa.Name,
			roleName,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return print(table)
}
