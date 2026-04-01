package get

import clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"

type natsUserRuleExecutor struct {
	namespace string
}

func newNatsUserRuleExecutor(namespace string) *natsUserRuleExecutor {
	return &natsUserRuleExecutor{namespace: namespace}
}

func (exe *natsUserRuleExecutor) GetName() string { return "" }

func (exe *natsUserRuleExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateNatsUserRulesOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateNatsUserRulesOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	response, err := clt.ListNatsUserRules()
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
