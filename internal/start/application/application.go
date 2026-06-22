package application

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type Options struct {
	Namespace string
	Name      string
}

type executor struct {
	namespace string
	name      string
}

func NewExecutor(opt Options) (exe execute.Executor) {
	return &executor{
		name:      opt.Name,
		namespace: opt.Namespace,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() (err error) {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	application, err := clt.GetApplicationByName(exe.name)
	if err != nil {
		return err
	}

	_, err = clt.StartApplication(application.Name)

	return
}
