package resource

type ControlPlane interface {
	GetUser() IofogUser
	UpdateUserTokens(string, string) IofogUser
	GetControllers() []Controller
	GetController(string) (Controller, error)
	GetEndpoint() (string, error)
	GetTrustCA() string
	UpdateController(Controller) error
	AddController(Controller) error
	DeleteController(string) error
	Sanitize() error
	Clone() ControlPlane
}
