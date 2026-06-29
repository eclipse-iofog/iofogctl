package exec

import (
	"errors"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

const (
	ErrMsgExecSessionQuotaExceeded = "Maximum of 3 concurrent exec sessions allowed for this microservice."
	ErrMsgTimeoutWaitingForAgent   = "Timeout waiting for agent connection. Please ensure the microservice/agent is running and try again."
	ErrMsgAuthenticationFailed     = "Authentication failed. Please check your credentials and try again."
	ErrMsgMicroserviceNotRunning   = "Microservice is not running. Please start the microservice first."
	ErrMsgDebugMicroserviceMissing = "Debug microservice not found. Run attach exec agent first to provision fog debug exec."
	ErrMsgDebugContainerTimeout    = "Timeout waiting for debug container to start. Please ensure the Agent is running and try again."
	ErrMsgAgentNotRunning          = "Agent is not running. Start the Agent before opening an exec session."
	ErrMsgInsufficientPermissions  = "Insufficient permissions. Required roles: SRE for Node Exec or Developer for Microservice Exec."
	ErrMsgOnlySREAccess            = "Only SRE can access system microservices. Please contact your administrator."
	ErrMsgExecRelayUnavailable     = "Exec session relay unavailable for cross-replica HA. Retry or check Controller NATS/AMQP relay connectivity."
	ErrMsgExecServerDraining       = "Server draining; retry during rollout or connect to another Controller replica."
	ErrMsgConnectionLost           = "Connection lost unexpectedly"
	ErrMsgMessageTooLarge          = "Message too large"
	ErrMsgServerError              = "Server error occurred"
	ErrMsgFailedToConnect          = "Failed to connect to exec session"
	ErrMsgConnectionClosed         = "Connection was closed"
)

func formatExecError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, client.ErrExecSessionQuotaExceeded):
		return ErrMsgExecSessionQuotaExceeded
	case errors.Is(err, client.ErrWsAgentTimeout):
		return ErrMsgTimeoutWaitingForAgent
	case errors.Is(err, client.ErrMicroserviceNotRunning):
		return ErrMsgMicroserviceNotRunning
	case errors.Is(err, client.ErrWsRelayUnavailable):
		return ErrMsgExecRelayUnavailable
	case errors.Is(err, client.ErrWsServerDraining):
		return ErrMsgExecServerDraining
	}

	errStr := err.Error()

	if strings.Contains(errStr, "Authentication failed") {
		return ErrMsgAuthenticationFailed
	}
	if strings.Contains(errStr, "Insufficient permissions") {
		return ErrMsgInsufficientPermissions
	}
	if strings.Contains(errStr, "Only SRE can access system microservices") {
		return ErrMsgOnlySREAccess
	}
	if strings.Contains(errStr, "Maximum of 3 concurrent exec sessions") {
		return ErrMsgExecSessionQuotaExceeded
	}
	if strings.Contains(errStr, "Timeout waiting for agent connection") {
		return ErrMsgTimeoutWaitingForAgent
	}
	if strings.Contains(errStr, "not running") || strings.Contains(errStr, "Not running") {
		return ErrMsgMicroserviceNotRunning
	}
	if strings.Contains(errStr, "close 1006") {
		return ErrMsgConnectionLost
	}
	if strings.Contains(errStr, "close 1009") {
		return ErrMsgMessageTooLarge
	}
	if strings.Contains(errStr, "close 1011") {
		return ErrMsgServerError
	}
	if strings.Contains(errStr, "exec WebSocket upgrade failed") || strings.Contains(errStr, "exec WebSocket dial") {
		return ErrMsgFailedToConnect
	}
	if strings.Contains(errStr, "use of closed network connection") {
		return ErrMsgConnectionClosed
	}
	if strings.Contains(errStr, execCloseCodeString(1000)) {
		return ""
	}

	if reason := extractCloseReason(errStr); reason != "" {
		return reason
	}

	return errStr
}

func extractCloseReason(errStr string) string {
	if idx := strings.Index(errStr, "reason:"); idx != -1 {
		return strings.TrimSuffix(strings.TrimSpace(errStr[idx+7:]), ".")
	}
	if idx := strings.Index(errStr, "exec WebSocket closed (code"); idx != -1 {
		if colon := strings.LastIndex(errStr, ": "); colon != -1 {
			return strings.TrimSpace(errStr[colon+2:])
		}
	}
	return ""
}
