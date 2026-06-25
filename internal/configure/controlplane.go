package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type kubernetesConfig struct {
	kubeConfig string
}

type caConfig struct {
	caFile string
	caB64  string
}

type controlPlaneExecutor struct {
	namespace        string
	kubernetesConfig kubernetesConfig
	caConfig         caConfig
	name             string
	remoteConfig     remoteConfig
}

func newControlPlaneExecutor(opt *Options) *controlPlaneExecutor {
	return &controlPlaneExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		remoteConfig: remoteConfig{
			keyFile: opt.KeyFile,
			user:    opt.User,
			port:    opt.Port,
		},
		kubernetesConfig: kubernetesConfig{
			kubeConfig: opt.KubeConfig,
		},
		caConfig: caConfig{
			caFile: opt.CAFile,
			caB64:  opt.CAB64,
		},
	}
}

func (exe *controlPlaneExecutor) GetName() string {
	return exe.name
}

func (exe *controlPlaneExecutor) Execute() error {
	caBase64, err := trust.NormalizeTrustCA(exe.caConfig.caFile, exe.caConfig.caB64)
	if err != nil {
		return err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	switch controlPlane := baseControlPlane.(type) {
	case *rsc.RemoteControlPlane:
		if exe.kubernetesConfig.kubeConfig != "" {
			return util.NewInputError("Cannot edit kube config of a Remote Control Plane")
		}
		if (remoteConfig{}) != exe.remoteConfig {
			return util.NewInputError("Cannot configure SSH settings on a Control Plane; use configure controller or configure agents")
		}
		if caBase64 == "" {
			return util.NewInputError("Nothing to configure for Remote Control Plane")
		}
		if err := rsc.SetTrustCA(controlPlane, caBase64); err != nil {
			return err
		}
		if err := trust.StoreCA(exe.namespace, caBase64); err != nil {
			return err
		}

	case *rsc.KubernetesControlPlane:
		if err := exe.kubernetesConfigure(controlPlane); err != nil {
			return err
		}
		if caBase64 != "" {
			if err := rsc.SetTrustCA(controlPlane, caBase64); err != nil {
				return err
			}
			if err := trust.StoreCA(exe.namespace, caBase64); err != nil {
				return err
			}
		}

	case *rsc.LocalControlPlane:
		if exe.kubernetesConfig.kubeConfig != "" || (remoteConfig{}) != exe.remoteConfig {
			return util.NewInputError("Cannot configure kube or SSH settings on a Local Control Plane")
		}
		if caBase64 == "" {
			return util.NewInputError("Nothing to configure for Local Control Plane")
		}
		if err := rsc.SetTrustCA(controlPlane, caBase64); err != nil {
			return err
		}
		if err := trust.StoreCA(exe.namespace, caBase64); err != nil {
			return err
		}
	}

	ns.SetControlPlane(baseControlPlane)
	return config.Flush()
}

func (exe *controlPlaneExecutor) kubernetesConfigure(controlPlane *rsc.KubernetesControlPlane) (err error) {
	if (remoteConfig{}) != exe.remoteConfig {
		return util.NewInputError("Cannot edit remote config of a Kubernetes Control Plane")
	}

	if exe.kubernetesConfig.kubeConfig != "" {
		controlPlane.KubeConfig = exe.kubernetesConfig.kubeConfig
	}

	return controlPlane.Sanitize()
}
