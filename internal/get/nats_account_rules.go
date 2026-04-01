package get

import clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"

type natsAccountRuleExecutor struct {
	namespace string
}

func newNatsAccountRuleExecutor(namespace string) *natsAccountRuleExecutor {
	return &natsAccountRuleExecutor{namespace: namespace}
}

func (exe *natsAccountRuleExecutor) GetName() string { return "" }

func (exe *natsAccountRuleExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateNatsAccountRulesOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateNatsAccountRulesOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	response, err := clt.ListNatsAccountRules()
	if err != nil {
		return nil, err
	}

	table := make([][]string, len(response.Rules)+1)
	table[0] = []string{"NAME"}
	for idx, rule := range response.Rules {
		table[idx+1] = []string{rule.Name}
	}
	return table, nil
}
