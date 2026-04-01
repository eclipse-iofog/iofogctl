package get

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type natsUserExecutor struct {
	namespace string
}

func newNatsUserExecutor(namespace string) *natsUserExecutor {
	return &natsUserExecutor{namespace: namespace}
}

func (exe *natsUserExecutor) GetName() string { return "" }

func (exe *natsUserExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateNatsUsersOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateNatsUsersOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	response, err := clt.ListNatsUsers()
	if err != nil {
		return nil, err
	}

	table := make([][]string, len(response.Users)+1)
	table[0] = []string{"NAME", "ACCOUNT", "APP", "BEARER", "PUBLIC KEY"}
	for idx, user := range response.Users {
		table[idx+1] = []string{
			user.Name,
			user.AccountName,
			user.ApplicationName,
			fmt.Sprintf("%t", user.IsBearer),
			user.PublicKey,
		}
	}
	return table, nil
}
