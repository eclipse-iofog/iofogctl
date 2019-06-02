package logs

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	return a
}

func (ns *agentExecutor) Execute() error {
	return nil
}
