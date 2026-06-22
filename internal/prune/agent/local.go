package pruneagent

import (
	"fmt"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) localAgentPrune(agent *rsc.LocalAgent) error {
	cfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), agent.Config, agent.Package)
	edgelet, err := install.NewLocalEdgelet(agent.Name, agent.UUID, cfg)
	if err != nil {
		return err
	}
	if err := edgelet.Prune(); err != nil {
		return util.NewInternalError(fmt.Sprintf("Could not prune local edgelet. Error: %s\n", err.Error()))
	}
	return nil
}
