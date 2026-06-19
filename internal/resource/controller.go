package resource

type Controller interface {
	GetName() string
	GetEndpoint() string
	GetCreatedTime() string
	SetName(string)
	Sanitize() error
	Clone() Controller
}
