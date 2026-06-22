package install

import (
	"fmt"
	"regexp"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	// TODO(v3.8.0): remove Moby client after local and remote control plane no longer use Go container deploy.
	dockengine "github.com/eclipse-iofog/iofogctl/pkg/containerengine/docker"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

type LocalControllerConfig struct {
	ContainerMap  map[string]*LocalContainerConfig
	Database      Database
	PidBaseDir    string
	EcnViewerPort int
	EcnViewerURL  string
	LogLevel      string
	Auth          Auth
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

func GetLocalContainerName(kind string, _ bool) string {
	if kind != "controller" {
		return ""
	}
	return sanitizeContainerName("iofog-controller")
}

func sanitizeContainerName(name string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	return r.ReplaceAllString(name, "-")
}

// LocalSystemImages optionally sets Router and Nats images and NATS enabling for the local controller.
type LocalSystemImages struct {
	Router      string
	Nats        string
	NatsEnabled *bool // nil = default enabled
}

// NewLocalControllerConfig generates a static controller config
func NewLocalControllerConfig(image string, credentials Credentials, auth Auth, db Database, events Events, systemImages *LocalSystemImages) *LocalContainerConfig {
	if image == "" {
		image = util.GetControllerImage()
	}

	sslValue := "false"
	if db.SSL != nil {
		sslValue = fmt.Sprintf("%t", *db.SSL)
	}

	caValue := ""
	if db.CA != nil {
		caValue = *db.CA
	}

	envs := []string{
		"CONTROL_PLANE=Remote",
		"DB_PROVIDER=" + db.Provider,
		"DB_HOST=" + db.Host,
		"DB_USERNAME=" + db.User,
		"DB_PASSWORD=" + db.Password,
		"DB_PORT=" + fmt.Sprintf("%d", db.Port),
		"DB_NAME=" + db.DatabaseName,
		"DB_USE_SSL=" + sslValue,
		"DB_SSL_CA=" + caValue,
	}
	_ = auth

	if events.AuditEnabled != nil {
		envs = append(envs, fmt.Sprintf("EVENT_AUDIT_ENABLED=%t", *events.AuditEnabled))
		if *events.AuditEnabled {
			if events.RetentionDays != 0 {
				envs = append(envs, fmt.Sprintf("EVENT_RETENTION_DAYS=%d", events.RetentionDays))
			}
			if events.CleanupInterval != 0 {
				envs = append(envs, fmt.Sprintf("EVENT_CLEANUP_INTERVAL=%d", events.CleanupInterval))
			}
			if events.CaptureIpAddress != nil {
				envs = append(envs, fmt.Sprintf("EVENT_CAPTURE_IP_ADDRESS=%t", *events.CaptureIpAddress))
			}
		}
	}

	if systemImages != nil {
		if systemImages.Router != "" {
			envs = append(envs, "ROUTER_IMAGE_1="+systemImages.Router, "ROUTER_IMAGE_2="+systemImages.Router)
		}
		natsImg := systemImages.Nats
		if natsImg == "" {
			natsImg = util.GetNatsImage()
		}
		if natsImg != "" {
			envs = append(envs, "NATS_IMAGE_1="+natsImg, "NATS_IMAGE_2="+natsImg)
		}
		natsEnabled := true
		if systemImages.NatsEnabled != nil {
			natsEnabled = *systemImages.NatsEnabled
		}
		envs = append(envs, fmt.Sprintf("NATS_ENABLED=%t", natsEnabled))
	}

	return &LocalContainerConfig{
		Host: "0.0.0.0",
		Ports: []port{
			{Host: iofog.ControllerPortString, Container: &LocalContainerPort{Port: iofog.ControllerPortString, Protocol: "tcp"}},
			{Host: iofog.ControllerHostECNViewerPortString, Container: &LocalContainerPort{Port: iofog.DefaultHTTPPortString, Protocol: "tcp"}},
		},
		ContainerName: GetLocalContainerName("controller", false),
		Image:         image,
		Privileged:    false,
		Binds: []string{
			"iofog-controller-db:/home/runner/.npm-global/lib/node_modules/@eclipse-iofog/iofogcontroller/src/data/sqlite_files/:rw",
			"iofog-controller-logs:/var/log/iofog-controller:rw",
		},
		Envs:        envs,
		NetworkMode: "bridge",
		Credentials: credentials,
	}
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

// Returns endpoint to reach controller container from within another container
func (lc *LocalContainer) GetLocalControllerEndpoint() (controllerEndpoint string, err error) {
	host, err := lc.GetContainerIP(GetLocalContainerName("controller", false))
	if err != nil {
		return controllerEndpoint, err
	}
	controllerEndpoint = fmt.Sprintf("http://%s:%s", host, iofog.ControllerPortString)
	return
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
