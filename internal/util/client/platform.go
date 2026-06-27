package client

import (
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

// PlatformWaitTimeout matches test/func/wait.bash microservice wait budget (20 × 20s).
const PlatformWaitTimeout = 400 * time.Second

func isAgentPlatformFailed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "platform reconcile failed")
}

func isServiceProvisioningFailed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "provisioning failed")
}

// WaitForAgentPlatformReadyWithRetry waits for fog platform Ready; on Failed, ReconcileAgent once and re-wait.
func WaitForAgentPlatformReadyWithRetry(namespace, uuid string) error {
	return ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
		err := clt.WaitForAgentPlatformReady(uuid, PlatformWaitTimeout)
		if err == nil {
			return nil
		}
		if !isAgentPlatformFailed(err) {
			return err
		}
		if _, recErr := clt.ReconcileAgent(uuid); recErr != nil {
			return err
		}
		return clt.WaitForAgentPlatformReady(uuid, PlatformWaitTimeout)
	})
}

// WaitForServiceProvisioningReadyWithRetry waits for hub provisioning ready; on Failed, ReconcileService once and re-wait.
func WaitForServiceProvisioningReadyWithRetry(namespace, name string) error {
	return ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
		err := clt.WaitForServiceProvisioningReady(name, PlatformWaitTimeout)
		if err == nil {
			return nil
		}
		if !isServiceProvisioningFailed(err) {
			return err
		}
		if _, recErr := clt.ReconcileService(name); recErr != nil {
			return err
		}
		return clt.WaitForServiceProvisioningReady(name, PlatformWaitTimeout)
	})
}

// ReconcileAgentByName resolves an agent name and enqueues platform reconcile.
func ReconcileAgentByName(namespace, name string) error {
	return ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
		agent, err := clt.GetAgentByName(name)
		if err != nil {
			return err
		}
		_, err = clt.ReconcileAgent(agent.UUID)
		return err
	})
}

// ReconcileAgentByNameAndWait reconciles an agent platform and waits until Ready.
func ReconcileAgentByNameAndWait(namespace, name string) error {
	var uuid string
	err := ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
		agent, err := clt.GetAgentByName(name)
		if err != nil {
			return err
		}
		uuid = agent.UUID
		_, err = clt.ReconcileAgent(uuid)
		return err
	})
	if err != nil {
		return err
	}
	return WaitForAgentPlatformReadyWithRetry(namespace, uuid)
}

// ReconcileServiceByName enqueues service hub reconcile.
func ReconcileServiceByName(namespace, name string) error {
	return ExecuteWithAuthRetry(namespace, func(clt *client.Client) error {
		_, err := clt.ReconcileService(name)
		return err
	})
}

// ReconcileServiceByNameAndWait reconciles a service hub and waits until ready.
func ReconcileServiceByNameAndWait(namespace, name string) error {
	if err := ReconcileServiceByName(namespace, name); err != nil {
		return err
	}
	return WaitForServiceProvisioningReadyWithRetry(namespace, name)
}
