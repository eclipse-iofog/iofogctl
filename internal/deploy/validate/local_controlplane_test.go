package validate

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

func requireInputError(t *testing.T, err error) {
	t.Helper()
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}

func TestLocalControlPlaneDeploy_AllowsInitial(t *testing.T) {
	require.NoError(t, LocalControlPlaneDeploy(1, false))
}

func TestLocalControlPlaneDeploy_AllowsRedeployLocalControlPlane(t *testing.T) {
	require.NoError(t, LocalControlPlaneDeploy(1, false))
}

func TestLocalControlPlaneDeploy_RejectsMultipleInFile(t *testing.T) {
	err := LocalControlPlaneDeploy(2, false)
	requireInputError(t, err)
	require.Contains(t, err.Error(), "multiple Local Control Planes")
}

func TestLocalControlPlaneDeploy_RejectsOtherControlPlaneKind(t *testing.T) {
	err := LocalControlPlaneDeploy(1, true)
	requireInputError(t, err)
	require.Contains(t, err.Error(), "different Control Plane kind")
}
