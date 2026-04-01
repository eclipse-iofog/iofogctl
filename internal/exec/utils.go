/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package exec

import (
	"strings"
)

// Error message constants to avoid duplication
const (
	ErrMsgAnotherUserConnected    = "Another user is already connected to this microservice. Only one user can connect at a time."
	ErrMsgTimeoutWaitingForAgent  = "Timeout waiting for agent connection. Please ensure the microservice/agent is running and try again."
	ErrMsgAuthenticationFailed    = "Authentication failed. Please check your credentials and try again."
	ErrMsgMicroserviceNotRunning  = "Microservice is not running. Please start the microservice first."
	ErrMsgExecNotEnabled          = "Microservice exec is not enabled. Please enable exec for this microservice."
	ErrMsgNoAvailableExecSession  = "No available exec session for this agent or microservice. Be sure to attach/link exec session to the agent or microservice first. If you already attached/linked exec session to the agent or microservice, please wait for the exec session to be ready."
	ErrMsgInsufficientPermissions = "Insufficient permissions. Required roles: SRE for Node Exec or Developer for Microservice Exec."
	ErrMsgOnlySREAccess           = "Only SRE can access system microservices. Please contact your administrator."
	ErrMsgConnectionLost          = "Connection lost unexpectedly"
	ErrMsgMessageTooLarge         = "Message too large"
	ErrMsgServerError             = "Server error occurred"
	ErrMsgFailedToConnect         = "Failed to connect to server"
	ErrMsgConnectionClosed        = "Connection was closed"
)

// formatWebSocketError formats WebSocket errors for better user experience
func formatWebSocketError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Handle specific WebSocket error patterns
	if strings.Contains(errStr, "close 1008") {
		// Extract the reason from the error message
		reason := extractCloseReason(errStr)

		// Try direct string matching first (more reliable)
		if strings.Contains(errStr, "No available exec session") {
			return ErrMsgNoAvailableExecSession
		}
		if strings.Contains(errStr, "Microservice has already active exec session") {
			return ErrMsgAnotherUserConnected
		}
		if strings.Contains(errStr, "Timeout waiting for agent connection") {
			return ErrMsgTimeoutWaitingForAgent
		}
		if strings.Contains(errStr, "Authentication failed") {
			return ErrMsgAuthenticationFailed
		}
		if strings.Contains(errStr, "Microservice is not running") {
			return ErrMsgMicroserviceNotRunning
		}
		if strings.Contains(errStr, "Microservice exec is not enabled") {
			return ErrMsgExecNotEnabled
		}
		if strings.Contains(errStr, "Microservice already has an active session") {
			return ErrMsgAnotherUserConnected
		}
		if strings.Contains(errStr, "Insufficient permissions") {
			return ErrMsgInsufficientPermissions
		}
		if strings.Contains(errStr, "Only SRE can access system microservices") {
			return ErrMsgOnlySREAccess
		}

		// If no direct match found, try the extracted reason
		if reason != "" {
			return reason
		}

		// Default fallback for unknown 1008 errors
		return "Policy violation: Access denied"
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

	if strings.Contains(errStr, "failed to connect") {
		return ErrMsgFailedToConnect
	}

	if strings.Contains(errStr, "use of closed network connection") {
		return ErrMsgConnectionClosed
	}

	// Default case - return the original error but clean it up
	if strings.Contains(errStr, "websocket: close") {
		// Extract the reason part if available
		if idx := strings.Index(errStr, "reason:"); idx != -1 {
			reason := strings.TrimSpace(errStr[idx+7:])
			return reason
		}
		// Extract the code and basic message
		if strings.Contains(errStr, "failed to read message:") {
			parts := strings.Split(errStr, "failed to read message:")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return errStr
}

// extractCloseReason extracts the reason from WebSocket close error messages
func extractCloseReason(errStr string) string {
	// Look for "reason:" pattern
	if idx := strings.Index(errStr, "reason:"); idx != -1 {
		reason := strings.TrimSpace(errStr[idx+7:])
		// Remove trailing period if present
		if strings.HasSuffix(reason, ".") {
			reason = reason[:len(reason)-1]
		}
		return reason
	}

	// Look for "policy violation:" pattern
	if idx := strings.Index(errStr, "policy violation:"); idx != -1 {
		reason := strings.TrimSpace(errStr[idx+18:])
		// Remove trailing period if present
		if strings.HasSuffix(reason, ".") {
			reason = reason[:len(reason)-1]
		}
		return reason
	}

	// Look for quoted reason at the end
	if strings.Contains(errStr, "close 1008") {
		// Try to extract the last quoted string
		parts := strings.Split(errStr, `"`)
		if len(parts) >= 2 {
			lastPart := parts[len(parts)-2] // Get the second-to-last part (the quoted reason)
			if lastPart != "" {
				return lastPart
			}
		}

		// Try to extract after "close 1008"
		if idx := strings.Index(errStr, "close 1008"); idx != -1 {
			afterClose := strings.TrimSpace(errStr[idx+10:])
			// Remove parentheses and other formatting
			afterClose = strings.TrimPrefix(afterClose, "(")
			afterClose = strings.TrimSuffix(afterClose, ")")
			afterClose = strings.TrimSpace(afterClose)

			// If it starts with a quote, extract the quoted part
			if strings.HasPrefix(afterClose, `"`) {
				if endIdx := strings.Index(afterClose[1:], `"`); endIdx != -1 {
					return afterClose[1 : endIdx+1]
				}
			}

			// If it contains a colon, extract after the colon
			if colonIdx := strings.Index(afterClose, ":"); colonIdx != -1 {
				reason := strings.TrimSpace(afterClose[colonIdx+1:])
				if strings.HasSuffix(reason, ".") {
					reason = reason[:len(reason)-1]
				}
				return reason
			}

			// Return the whole thing if it looks like a reason
			if len(afterClose) > 0 && !strings.Contains(afterClose, "websocket") {
				return afterClose
			}
		}
	}

	return ""
}
