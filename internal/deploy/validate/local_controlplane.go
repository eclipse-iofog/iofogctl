package validate

import "github.com/eclipse-iofog/iofogctl/pkg/util"

func LocalControlPlaneDeploy(localCPCount int, namespaceHasOtherControlPlane bool) error {
	if localCPCount > 1 {
		return util.NewInputError("Specified multiple Local Control Planes in a single deploy file")
	}
	if localCPCount > 0 && namespaceHasOtherControlPlane {
		return util.NewInputError(
			"Namespace already has a different Control Plane kind; delete it before deploying LocalControlPlane",
		)
	}
	return nil
}
