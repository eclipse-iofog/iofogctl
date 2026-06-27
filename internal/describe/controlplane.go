package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controlPlaneExecutor struct {
	namespace string
	filename  string
}

func newControlPlaneExecutor(namespace, filename string) *controlPlaneExecutor {
	return &controlPlaneExecutor{
		namespace: namespace,
		filename:  filename,
	}
}

func (exe *controlPlaneExecutor) GetName() string {
	return exe.namespace
}

func (exe *controlPlaneExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Generate header
	var header config.Header
	switch controlPlane := baseControlPlane.(type) {
	case *rsc.KubernetesControlPlane:
		header = exe.generateControlPlaneHeader(config.KubernetesControlPlaneKind, controlPlane)
	case *rsc.RemoteControlPlane:
		header = exe.generateControlPlaneHeader(config.RemoteControlPlaneKind, controlPlane)
	case *rsc.LocalControlPlane:
		header = exe.generateControlPlaneHeader(config.LocalControlPlaneKind, controlPlane)
	default:
		return util.NewInternalError("Could not convert Control Plane to dynamic type")
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}

func (exe *controlPlaneExecutor) generateControlPlaneHeader(kind config.Kind, controlPlane rsc.ControlPlane) config.Header {
	return config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       kind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      "controlPlane",
		},
		Spec: controlPlane,
	}
}
