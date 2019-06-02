package logs

type microserviceExecutor struct {
	namespace string
	name      string
}

func newMicroserviceExecutor(namespace, name string) *microserviceExecutor {
	m := &microserviceExecutor{}
	m.namespace = namespace
	m.name = name
	return m
}

func (ns *microserviceExecutor) Execute() error {
	return nil
}
