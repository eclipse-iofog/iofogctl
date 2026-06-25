package deletelocalcontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
}

func NewExecutor(namespace string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
	}
	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return "Delete Control Plane"
}

func (exe *Executor) Execute() (err error) {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	controlPlane, ok := baseControlPlane.(*rsc.LocalControlPlane)
	if !ok {
		return util.NewError("Could not convert Control Plane to Local Control Plane")
	}

	name := ControlPlaneSystemAgentName(controlPlane)

	if controlPlane.SystemAgent != nil {
		if err := teardownEdgeletControlPlane(exe.namespace, controlPlane, name); err != nil {
			return err
		}
	}

	ns.DeleteControlPlane()
	return config.Flush()
}

// IsLocalEdgeletControlPlane reports whether the namespace uses edgelet host teardown.
func IsLocalEdgeletControlPlane(cp rsc.ControlPlane) bool {
	return isLocalEdgeletControlPlane(cp)
}
