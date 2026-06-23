package deploylocalcontrolplane

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func resolveLocalControlPlaneEndpoint(cp *rsc.LocalControlPlane) string {
	if cp.Endpoint != "" {
		return cp.Endpoint
	}
	if cp.Controller.PublicUrl != "" {
		return cp.Controller.PublicUrl
	}
	return fmt.Sprintf("http://localhost:%s", iofog.ControllerPortString)
}

// BuildLocalEdgelet constructs a LocalEdgelet from LocalControlPlane systemAgent settings.
func BuildLocalEdgelet(cp *rsc.LocalControlPlane, name, agentUUID string) (*install.LocalEdgelet, error) {
	sys := cp.SystemAgent
	var cfg *rsc.AgentConfiguration
	var pkg rsc.Package
	var scripts *rsc.AgentScripts
	if sys != nil {
		cfg = sys.AgentConfiguration
		pkg = sys.Package
		scripts = sys.Scripts
	}

	cfg = deployairgap.EnsureAgentConfig(cfg)
	deployairgap.ResolveAgentDeployment(cfg, pkg.Container.Image)

	installCfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), cfg, pkg)
	edgelet, err := install.NewLocalEdgelet(name, agentUUID, installCfg)
	if err != nil {
		return nil, err
	}

	if scripts != nil {
		procs := install.EdgeletProcedures{AgentProcedures: scripts.AgentProcedures}
		if err := edgelet.CustomizeProcedures(scripts.Directory, &procs); err != nil {
			return nil, err
		}
	}
	if pkg.Container.Image != "" {
		if err := edgelet.SetContainerImage(pkg.Container.Image); err != nil {
			return nil, err
		}
	} else if pkg.Version != "" {
		if err := edgelet.SetVersion(pkg.Version); err != nil {
			return nil, err
		}
	}
	return edgelet, nil
}

// EdgeletHostTeardown runs host-level edgelet teardown commands.
type EdgeletHostTeardown interface {
	Deprovision() error
	DeleteControlPlane() error
	Uninstall(removeData bool) error
}

// BuildEdgeletForTeardown selects LocalEdgelet or RemoteEdgelet for control plane host teardown.
func BuildEdgeletForTeardown(cp *rsc.LocalControlPlane, namespace, name, agentUUID string) (EdgeletHostTeardown, error) {
	if namespace != "" {
		ns, err := config.GetNamespace(namespace)
		if err == nil {
			if baseAgent, err := ns.GetAgent(name); err == nil {
				if remote, ok := baseAgent.(*rsc.RemoteAgent); ok && remote.ValidateSSH() == nil {
					return buildRemoteEdgeletForTeardown(remote)
				}
			}
		}
	}
	return BuildLocalEdgelet(cp, name, agentUUID)
}

func buildRemoteEdgeletForTeardown(agent *rsc.RemoteAgent) (*install.RemoteEdgelet, error) {
	cfg := deployairgap.EdgeletInstallConfig("linux", agent.Config, agent.Package)
	edgelet, err := install.NewRemoteEdgelet(
		agent.SSH.User,
		agent.Host,
		agent.SSH.Port,
		agent.SSH.KeyFile,
		agent.Name,
		agent.UUID,
		cfg,
	)
	if err != nil {
		return nil, err
	}
	if agent.Scripts != nil {
		procs := install.EdgeletProcedures{AgentProcedures: agent.Scripts.AgentProcedures}
		if err := edgelet.CustomizeProcedures(agent.Scripts.Directory, &procs); err != nil {
			return nil, err
		}
	}
	if agent.Package.Container.Image != "" {
		if err := edgelet.SetContainerImage(agent.Package.Container.Image); err != nil {
			return nil, err
		}
	} else if agent.Package.Version != "" {
		if err := edgelet.SetVersion(agent.Package.Version); err != nil {
			return nil, err
		}
	}
	return edgelet, nil
}

func installHostEdgelet(cp *rsc.LocalControlPlane, name string) (*install.LocalEdgelet, error) {
	edgelet, err := BuildLocalEdgelet(cp, name, "")
	if err != nil {
		return nil, err
	}

	util.SpinStart("Installing edgelet")
	if err := edgelet.Bootstrap(); err != nil {
		return nil, err
	}
	return edgelet, nil
}

func deployPrivateEdgeletRegistry(cp *rsc.LocalControlPlane, edgelet *install.LocalEdgelet, opts TranslateOptions) (*int, error) {
	if !NeedsPrivateEdgeletRegistry(cp) {
		return nil, nil
	}

	result, err := TranslateLocalControlPlane(cp, opts)
	if err != nil {
		return nil, err
	}
	if len(result.Registry) == 0 {
		return nil, util.NewError("private registry translation produced no manifest")
	}

	path, cleanup, err := edgelet.WriteDeployManifest(result.Registry, "edgelet-registry")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if err := edgelet.DeployFromFile(path); err != nil {
		return nil, fmt.Errorf("edgelet registry deploy failed: %w", err)
	}

	output, err := edgelet.RegistryList()
	if err != nil {
		return nil, fmt.Errorf("edgelet registry ls failed: %w", err)
	}

	pkg := cp.Controller.Package
	id, err := install.ParseEdgeletRegistryID(output, pkg.Registry, pkg.Username)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func deployEdgeletControlPlane(cp *rsc.LocalControlPlane, edgelet *install.LocalEdgelet, opts TranslateOptions, registryID *int) error {
	opts.RegistryID = ResolveEdgeletRegistryID(cp, registryID)
	result, err := TranslateLocalControlPlane(cp, opts)
	if err != nil {
		return err
	}

	path, cleanup, err := edgelet.WriteDeployManifest(result.ControlPlane, "edgelet-controlplane")
	if err != nil {
		return err
	}
	defer cleanup()

	if err := edgelet.DeployFromFile(path); err != nil {
		return fmt.Errorf("edgelet control plane deploy failed: %w", err)
	}
	return nil
}

func deployControllerRegistry(namespace string, cp *rsc.LocalControlPlane) error {
	if !NeedsPrivateEdgeletRegistry(cp) {
		return nil
	}
	pkg := cp.Controller.Package
	if pkg == nil {
		return nil
	}

	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	url := pkg.Registry
	username := pkg.Username
	password := pkg.Password
	email := pkg.Email
	if email == "" {
		email = "registry@local"
	}

	createRequest := &client.RegistryCreateRequest{
		URL:      url,
		IsPublic: false,
		Username: username,
		Password: password,
		Email:    email,
	}
	if _, err = clt.CreateRegistry(createRequest); err != nil {
		return fmt.Errorf("failed to register private registry with controller: %w", err)
	}
	return nil
}

func persistLocalControllerStub(cp *rsc.LocalControlPlane, name, endpoint string) error {
	controller := &rsc.LocalController{
		Name:     name,
		Endpoint: endpoint,
		Created:  util.NowUTC(),
	}
	cp.Endpoint = endpoint
	return cp.UpdateController(controller)
}
