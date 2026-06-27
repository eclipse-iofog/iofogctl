package deployremotecontrolplane

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// BuildRemoteEdgelet constructs a RemoteEdgelet from per-controller systemAgent settings.
func BuildRemoteEdgelet(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, agentUUID string) (*install.RemoteEdgelet, error) {
	sys := ctrl.SystemAgent
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

	installCfg := deployairgap.EdgeletInstallConfig("linux", cfg, pkg)
	edgelet, err := install.NewRemoteEdgelet(
		ctrl.SSH.User,
		ctrl.Host,
		ctrl.SSH.Port,
		ctrl.SSH.KeyFile,
		ctrl.Name,
		agentUUID,
		installCfg,
	)
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

func isNonRoutableControllerHost(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return true
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	switch strings.ToLower(host) {
	case "0.0.0.0", "127.0.0.1", "localhost", "::1":
		return true
	default:
		return false
	}
}

func systemAgentConfigHost(ctrl *rsc.RemoteController) string {
	if ctrl == nil || ctrl.SystemAgent == nil || ctrl.SystemAgent.AgentConfiguration == nil {
		return ""
	}
	if host := ctrl.SystemAgent.AgentConfiguration.Host; host != nil {
		if h := strings.TrimSpace(*host); h != "" && !isNonRoutableControllerHost(h) {
			return h
		}
	}
	return ""
}

func resolveControllerAPIHost(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) string {
	if cp != nil {
		if publicURL := strings.TrimSpace(cp.Controller.PublicUrl); publicURL != "" {
			if u, err := util.GetBaseURL(publicURL); err == nil && u.Host != "" {
				return u.Host
			}
		}
	}
	if host := systemAgentConfigHost(ctrl); host != "" {
		return host
	}
	if !isNonRoutableControllerHost(ctrl.Host) {
		return strings.TrimSpace(ctrl.Host)
	}
	if cp != nil {
		if cpEndpoint := strings.TrimSpace(cp.Endpoint); cpEndpoint != "" {
			if u, err := util.GetBaseURL(cpEndpoint); err == nil && u.Host != "" && !isNonRoutableControllerHost(u.Hostname()) {
				return u.Host
			}
		}
	}
	return strings.TrimSpace(ctrl.Host)
}

func ResolveControllerHostEndpoint(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) (string, error) {
	if cp != nil {
		if publicURL := strings.TrimSpace(cp.Controller.PublicUrl); publicURL != "" {
			return publicURL, nil
		}
	}

	tls := EffectiveControllerTLS(cp, ctrl)
	useHTTPS := tls != nil && tls.Cert != "" && tls.Key != ""
	apiHost := resolveControllerAPIHost(cp, ctrl)
	return util.GetControllerEndpoint(apiHost, useHTTPS)
}

func DeployHostEdgelet(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, namespace string) (*install.RemoteEdgelet, error) {
	edgelet, err := BuildRemoteEdgelet(cp, ctrl, "")
	if err != nil {
		return nil, err
	}

	if cp.Airgap {
		if err := transferControllerHostAirgap(context.Background(), namespace, cp, ctrl, edgelet); err != nil {
			return nil, err
		}
	}

	util.SpinStart("Installing edgelet on " + ctrl.Name)
	if err := edgelet.Bootstrap(); err != nil {
		return nil, err
	}
	return edgelet, nil
}

func transferControllerHostAirgap(ctx context.Context, namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, edgelet *install.RemoteEdgelet) error {
	if ctrl.SystemAgent == nil || ctrl.SystemAgent.AgentConfiguration == nil {
		return util.NewInputError("systemAgent.config is required for airgap deployment on controller " + ctrl.Name)
	}

	isInitial, err := deployairgap.IsInitialDeployment(namespace)
	if err != nil {
		return fmt.Errorf("failed to determine deployment type: %w", err)
	}

	images, err := deployairgap.CollectControllerImages(namespace, cp, isInitial)
	if err != nil {
		return fmt.Errorf("failed to collect controller images: %w", err)
	}

	platform, err := deployairgap.ResolvePlatform(ctrl.SystemAgent.AgentConfiguration.Arch)
	if err != nil {
		return fmt.Errorf("controller %s: %w", ctrl.Name, err)
	}
	opts, err := deployairgap.ControllerAirgapLoadOptions(ctrl.SystemAgent.AgentConfiguration)
	if err != nil {
		return fmt.Errorf("controller %s: %w", ctrl.Name, err)
	}

	imageList := []string{images.Controller}
	for _, img := range []string{images.NatsAMD64, images.NatsARM64, images.NatsRISCV64, images.NatsARM} {
		if img != "" {
			imageList = append(imageList, img)
		}
	}

	if deployairgap.IsNativeDeployment(opts.DeploymentType) {
		remoteBinPath, err := deployairgap.TransferAgentAirgapBinary(ctx, namespace, ctrl.Host, &ctrl.SSH, platform)
		if err != nil {
			return fmt.Errorf("failed to transfer edgelet binary to %s: %w", ctrl.Name, err)
		}
		if err := edgelet.SetAirgap(remoteBinPath); err != nil {
			return fmt.Errorf("failed to configure edgelet airgap binary on %s: %w", ctrl.Name, err)
		}
	}

	if len(imageList) > 0 {
		if err := deployairgap.TransferAirgapImages(ctx, namespace, ctrl.Host, &ctrl.SSH, platform, opts, imageList); err != nil {
			return fmt.Errorf("failed to transfer images to controller %s: %w", ctrl.Name, err)
		}
	}
	return nil
}

func DeployPrivateEdgeletRegistry(cp *rsc.RemoteControlPlane, edgelet *install.RemoteEdgelet, opts TranslateOptions) (*int, error) {
	if !NeedsPrivateEdgeletRegistry(cp) {
		return nil, nil
	}

	result, err := TranslateRemoteControlPlane(cp, nil, opts)
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

func DeployEdgeletControlPlane(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, edgelet *install.RemoteEdgelet, opts TranslateOptions, registryID *int) error {
	opts.RegistryID = ResolveEdgeletRegistryID(cp, registryID)
	result, err := TranslateRemoteControlPlane(cp, ctrl, opts)
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

func deployControllerRegistry(namespace string, cp *rsc.RemoteControlPlane) error {
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

	email := pkg.Email
	if email == "" {
		email = "registry@local"
	}

	createRequest := &client.RegistryCreateRequest{
		URL:      pkg.Registry,
		IsPublic: false,
		Username: pkg.Username,
		Password: pkg.Password,
		Email:    email,
	}
	if _, err = clt.CreateRegistry(createRequest); err != nil {
		return fmt.Errorf("failed to register private registry with controller: %w", err)
	}
	return nil
}

func deployRemoteControlPlaneHost(namespace, name string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) error {
	if err := ctrl.ValidateSSH(); err != nil {
		return err
	}

	edgelet, err := DeployHostEdgelet(cp, ctrl, namespace)
	if err != nil {
		return err
	}

	translateOpts := TranslateOptions{
		Name:      name,
		Namespace: namespace,
	}

	registryID, err := DeployPrivateEdgeletRegistry(cp, edgelet, translateOpts)
	if err != nil {
		return err
	}

	if err := DeployEdgeletControlPlane(cp, ctrl, edgelet, translateOpts, registryID); err != nil {
		return err
	}

	endpoint, err := ResolveControllerHostEndpoint(cp, ctrl)
	if err != nil {
		return err
	}
	ctrl.Endpoint = endpoint
	ctrl.Created = util.NowUTC()
	return nil
}
