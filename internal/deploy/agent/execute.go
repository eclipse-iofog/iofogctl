package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
	IsSystem  bool
	Tags      *[]string
}

func NewRemoteExecutorYAML(opt Options) (exe execute.Executor, err error) {
	// Read the input file
	agent, err := rsc.UnmarshallRemoteAgent(opt.Yaml)
	if err != nil {
		return exe, err
	}

	if len(opt.Name) > 0 {
		agent.Name = opt.Name
	}

	// Validate
	if err = ValidateRemoteAgent(&agent); err != nil {
		return
	}

	remoteExe := newRemoteExecutor(opt.Namespace, &agent)
	return newFacadeExecutor(remoteExe, opt.Namespace, &agent, opt.IsSystem, opt.Tags), nil
}

func NewLocalExecutorYAML(opt Options) (exe execute.Executor, err error) {
	// Read the input file
	agent, err := rsc.UnmarshallLocalAgent(opt.Yaml)
	if err != nil {
		return exe, err
	}

	if len(opt.Name) > 0 {
		agent.Name = opt.Name
	}

	if err = rsc.ValidateAgentPackage("LocalAgent", agent.Package, agent.Config); err != nil {
		return exe, err
	}

	localExe, err := newLocalExecutor(opt.Namespace, &agent, opt.IsSystem)
	if err != nil {
		return nil, err
	}
	return newFacadeExecutor(localExe, opt.Namespace, &agent, opt.IsSystem, opt.Tags), nil
}
