package install

import (
	"regexp"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	dockengine "github.com/eclipse-iofog/iofogctl/pkg/containerengine/docker"
)

// LocalContainer struct to encapsulate utilities around docker
type LocalContainer struct {
	client *dockengine.Client
}

// ExecResult contains the output of a command ran into docker exec
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type port struct {
	Host      string
	Container *LocalContainerPort
}

type Credentials struct {
	User     string
	Password string
}

type LocalContainerConfig struct {
	Host          string
	Ports         []port
	ContainerName string
	Image         string
	DefaultImage  string
	Privileged    bool
	Binds         []string
	Envs          []string
	NetworkMode   string
	Credentials   Credentials
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

func sanitizeContainerName(name string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	return r.ReplaceAllString(name, "-")
}

// NewLocalContainerClient dials the container engine using ResolveContainerEngineURL.
func NewLocalContainerClient(engine string, agentCfg *client.AgentConfiguration) (*LocalContainer, error) {
	cli, err := NewContainerEngineClient(engine, agentCfg)
	if err != nil {
		return nil, err
	}
	return &LocalContainer{client: cli}, nil
}

// GetLogsByName returns the logs of the container specified by name
func (lc *LocalContainer) GetLogsByName(name string) (stdout, stderr string, err error) {
	return lc.client.GetLogsByName(name)
}

// GetContainerByName returns a container summary by name.
func (lc *LocalContainer) GetContainerByName(name string) (dockengine.ContainerSummary, error) {
	return lc.client.GetContainerByName(name)
}

func (lc *LocalContainer) ListContainers() ([]dockengine.ContainerSummary, error) {
	return lc.client.ListContainers(true)
}

// CleanContainer stops and remove a container based on a container name
func (lc *LocalContainer) CleanContainer(name string) error {
	return lc.client.RemoveContainerByName(name)
}

func (lc *LocalContainer) CleanContainerByID(id string) error {
	return lc.client.RemoveContainerByID(id)
}

// DeployContainer deploys a container based on an image and a port mapping
func (lc *LocalContainer) DeployContainer(containerConfig *LocalContainerConfig) (string, error) {
	portBindings := map[string]dockengine.PortBinding{}
	for _, p := range containerConfig.Ports {
		proto := p.Container.Protocol
		if proto == "" {
			proto = "tcp"
		}
		portBindings[p.Host] = dockengine.PortBinding{
			HostIP:        containerConfig.Host,
			HostPort:      p.Host,
			ContainerPort: p.Container.Port,
			Protocol:      proto,
		}
	}

	opts := dockengine.DeployOptions{
		Name:          containerConfig.ContainerName,
		Image:         containerConfig.Image,
		Env:           containerConfig.Envs,
		Binds:         containerConfig.Binds,
		Privileged:    containerConfig.Privileged,
		NetworkMode:   containerConfig.NetworkMode,
		PortBindings:  portBindings,
		RestartPolicy: dockengine.RestartPolicyAlways,
		StopTimeout:   60 * time.Second,
	}
	pullOpts := dockengine.PullOptions{
		Username: containerConfig.Credentials.User,
		Password: containerConfig.Credentials.Password,
	}
	return lc.client.DeployContainer(opts, pullOpts)
}

func (lc *LocalContainer) GetContainerIP(name string) (ip string, err error) {
	return lc.client.GetContainerIP(name)
}

func (lc *LocalContainer) WaitForCommand(containerName string, condition *regexp.Regexp, command ...string) error {
	return lc.client.WaitForCommand(containerName, func(stdout string) bool {
		return condition.MatchString(stdout)
	}, command...)
}

func (lc *LocalContainer) ExecuteCmd(name string, cmd []string) (execResult ExecResult, err error) {
	result, err := lc.client.ExecuteCmd(name, cmd)
	if err != nil {
		return execResult, err
	}
	return ExecResult{
		StdOut:   result.StdOut,
		StdErr:   result.StdErr,
		ExitCode: result.ExitCode,
	}, nil
}

func (lc *LocalContainer) CopyToContainer(name, source, dest string) (err error) {
	return lc.client.CopyToContainer(name, source, dest)
}

// EdgeletContainerName is the default docker container name for edgelet.
const EdgeletContainerName = "edgelet"

func edgeletAgentSDKConfig(cfg EdgeletInstallConfig) *client.AgentConfiguration {
	if cfg.Runtime == nil {
		return nil
	}
	return &cfg.Runtime.Agent
}

// NewLocalContainerClientFromEdgeletCfg dials the engine configured for an edgelet install.
func NewLocalContainerClientFromEdgeletCfg(cfg EdgeletInstallConfig) (*LocalContainer, error) {
	return NewLocalContainerClient(cfg.containerEngine(), edgeletAgentSDKConfig(cfg))
}
