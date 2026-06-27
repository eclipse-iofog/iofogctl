package execute

type Executor interface {
	Execute() error
	GetName() string
}

type ProvisioningExecutor interface {
	ProvisionAgent() (string, error)
}
