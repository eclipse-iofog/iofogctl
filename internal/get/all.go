package get

type allExecutor struct {
	namespace string
}

func newAllExecutor(namespace string) *allExecutor {
	exe := &allExecutor{}
	exe.namespace = namespace
	return exe
}

func (exe *allExecutor) Execute() error {
	if err := newControllerExecutor(exe.namespace).Execute(); err != nil {
		return err
	}
	if err := newAgentExecutor(exe.namespace).Execute(); err != nil {
		return err
	}
	if err := newMicroserviceExecutor(exe.namespace).Execute(); err != nil {
		return err
	}

	return nil
}
