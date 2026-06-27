package logs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func streamLogSession(session *client.LogSession) error {
	defer session.Close()

	for {
		frame, err := session.Read()
		if err != nil {
			return util.NewError(formatLogError(err))
		}
		if frame == nil {
			return nil
		}

		switch frame.Type {
		case client.LogMessageLine:
			writeLogLine(frame.Data)
		case client.LogMessageStart:
			continue
		case client.LogMessageStop:
			return nil
		case client.LogMessageError:
			errorMsg := string(frame.Data)
			if errorMsg == "" {
				errorMsg = "Log streaming error occurred"
			}
			return util.NewError(errorMsg)
		}
	}
}

func writeLogLine(data []byte) {
	if len(data) == 0 {
		util.WriteStdout([]byte{'\n'})
		return
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	util.WriteStdout(data)
}

func runRemoteLogStream(clt *client.Client, dial func(*client.Client) (*client.LogSession, error)) error {
	util.SpinHandlePrompt()
	session, err := dial(clt)
	if err != nil {
		util.SpinHandlePromptComplete()
		return util.NewError(formatLogError(err))
	}

	if err := streamLogSession(session); err != nil {
		util.SpinHandlePromptComplete()
		return err
	}

	util.SpinHandlePromptComplete()
	return nil
}

func formatLogError(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, client.ErrLogSessionUnavailable):
		return client.ErrLogSessionUnavailable.Error()
	case errors.Is(err, client.ErrLogAuthenticationFailed):
		return client.ErrLogAuthenticationFailed.Error()
	case errors.Is(err, client.ErrAgentNotRunning):
		return client.ErrAgentNotRunning.Error()
	case errors.Is(err, client.ErrMicroserviceNotRunning):
		return client.ErrMicroserviceNotRunning.Error()
	case errors.Is(err, client.ErrLogInsufficientPermissions):
		return client.ErrLogInsufficientPermissions.Error()
	case errors.Is(err, client.ErrLogPolicyViolation):
		return client.ErrLogPolicyViolation.Error()
	case errors.Is(err, client.ErrLogConnectionLost):
		return client.ErrLogConnectionLost.Error()
	case errors.Is(err, client.ErrLogMessageTooLarge):
		return client.ErrLogMessageTooLarge.Error()
	case errors.Is(err, client.ErrLogServerError):
		return client.ErrLogServerError.Error()
	}

	errStr := err.Error()
	if strings.Contains(errStr, "close 1008") {
		if strings.Contains(errStr, "No available log session") {
			return "No available log session"
		}
		if strings.Contains(errStr, "Authentication failed") {
			return "Authentication failed"
		}
		if strings.Contains(errStr, "Agent is not running") {
			return "Agent is not running"
		}
		if strings.Contains(errStr, "Microservice is not running") {
			return "Microservice is not running"
		}
		if strings.Contains(errStr, "Insufficient permissions") {
			return "Insufficient permissions"
		}
		return "Policy violation: Access denied"
	}
	if strings.Contains(errStr, "close 1006") {
		return "Connection lost"
	}
	if strings.Contains(errStr, "close 1009") {
		return "Message too large"
	}
	if strings.Contains(errStr, "close 1011") {
		return "Server error"
	}
	if strings.Contains(errStr, "failed to connect") {
		return "Failed to connect to log stream"
	}

	return fmt.Sprintf("Log stream error: %v", err)
}
