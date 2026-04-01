package get

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type natsAccountExecutor struct {
	namespace string
}

func newNatsAccountExecutor(namespace string) *natsAccountExecutor {
	return &natsAccountExecutor{namespace: namespace}
}

func (exe *natsAccountExecutor) GetName() string { return "" }

func (exe *natsAccountExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateNatsAccountsOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateNatsAccountsOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	response, err := clt.ListNatsAccounts()
	if err != nil {
		return nil, err
	}

	table := make([][]string, len(response.Accounts)+1)
	table[0] = []string{"NAME", "APP ID", "SYSTEM", "PUBLIC KEY"}
	for idx, account := range response.Accounts {
		table[idx+1] = []string{
			account.Name,
			fmt.Sprintf("%d", account.ApplicationID),
			fmt.Sprintf("%t", account.IsSystem),
			account.PublicKey,
		}
	}
	return table, nil
}
