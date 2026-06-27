package deleteall

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	deleteagents "github.com/eclipse-iofog/iofogctl/internal/delete/agents"
	deletecontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane"
	deletelocalcontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane/local"
	deletevolume "github.com/eclipse-iofog/iofogctl/internal/delete/volume"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace string, useDetached, force, deleteNamespace bool) error {
	// Make sure to update config despite failure
	defer config.Flush()

	// Get namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Delete Volumes
	if len(ns.Volumes) > 0 {
		util.SpinStart("Deleting Volumes")
		var executors []execute.Executor
		for _, volume := range ns.Volumes {
			exe, err := deletevolume.NewExecutor(namespace, volume.Name)
			if err != nil {
				return err
			}
			executors = append(executors, exe)
		}
		if err := runExecutors(executors); err != nil {
			return err
		}
	}

	if !useDetached {
		// Delete applications
		util.SpinStart("Deleting Applications")
		clt, err := clientutil.NewControllerClient(namespace)
		if err != nil {
			return err
		}

		applications, err := clt.GetAllApplications()
		if err != nil {
			return err
		}

		for _, application := range applications.Applications {
			if err := clt.DeleteApplication(application.Name); err != nil {
				return err
			}
		}
	}

	// Delete non-control-plane agents first while the controller is still reachable.
	var excludeAgentNames []string
	if cp, cpErr := ns.GetControlPlane(); cpErr == nil {
		if localCP, ok := cp.(*rsc.LocalControlPlane); ok && localCP.SystemAgent != nil {
			excludeAgentNames = []string{deletelocalcontrolplane.ControlPlaneSystemAgentName(localCP)}
		}
	}
	agentTargets, err := deleteagents.CollectDeleteTargets(ns, namespace, force, excludeAgentNames)
	if err != nil {
		return err
	}
	if len(agentTargets) > 0 {
		util.SpinStart("Deleting Agents")

		var executors []execute.Executor
		for _, target := range agentTargets {
			exe, err := deleteagent.NewExecutor(namespace, target.Name, useDetached, target.Force)
			if err != nil {
				return err
			}
			executors = append(executors, exe)
		}
		if err := runExecutors(executors); err != nil {
			return err
		}
	}

	if !useDetached {
		// Delete Controllers
		util.SpinStart("Deleting Control Plane ")
		exe, err := deletecontrolplane.NewExecutor(namespace, deleteNamespace)
		if err != nil {
			return err
		}
		if err := exe.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}
