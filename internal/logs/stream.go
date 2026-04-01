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

package logs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	ws "github.com/eclipse-iofog/iofogctl/internal/util/websocket"
)

// LogStream handles streaming logs from WebSocket connection
type LogStream struct {
	wsClient    *ws.Client
	ctx         context.Context
	cancel      context.CancelFunc
	cleanupOnce sync.Once
	stdoutMutex sync.Mutex
}

// NewLogStream creates a new LogStream handler
func NewLogStream(wsClient *ws.Client) *LogStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &LogStream{
		wsClient: wsClient,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (ls *LogStream) writeToStdout(data []byte) {
	ls.stdoutMutex.Lock()
	defer ls.stdoutMutex.Unlock()
	os.Stdout.Write(data)
	os.Stdout.Sync()
}

func (ls *LogStream) cleanup() {
	ls.cleanupOnce.Do(func() {
		if ls.wsClient != nil {
			ls.wsClient.Close()
		}
	})
}

// Start starts the log stream handler
func (ls *LogStream) Start() error {
	defer ls.cleanup()

	// Monitor log messages
	go func() {
		for {
			select {
			case <-ls.ctx.Done():
				return
			default:
				msg, err := ls.wsClient.ReadMessage()
				if err != nil {
					// Check if this is a normal closure
					if ls.wsClient.IsNormalClosure(err) {
						// Normal closure - don't report as error, just exit gracefully
						ls.cancel()
						return
					}
					// This is an actual error - format and return
					ls.cancel()
					return
				}
				if msg == nil {
					// Normal termination (no error, no message)
					ls.cancel()
					return
				}

				// Handle different message types
				if msg.IsLogLineMessage() {
					// Write log line to stdout
					// Ensure the data ends with a newline if it doesn't already
					data := msg.Data
					if len(data) == 0 {
						// Empty line - just write a newline
						data = []byte{'\n'}
					} else if data[len(data)-1] != '\n' {
						// Non-empty line without newline - add one
						data = append(data, '\n')
					}
					// If data already ends with newline, use it as-is
					ls.writeToStdout(data)
				} else if msg.IsLogStartMessage() {
					// Log streaming started - can parse sessionId from data if needed
					// For now, just continue
				} else if msg.IsLogStopMessage() {
					// Log streaming stopped - exit gracefully
					ls.cancel()
					return
				} else if msg.IsLogErrorMessage() {
					// Log streaming error - write error and exit
					errorMsg := string(msg.Data)
					if errorMsg == "" {
						errorMsg = "Log streaming error occurred"
					}
					ls.writeToStdout([]byte(fmt.Sprintf("Error: %s\n", errorMsg)))
					ls.cancel()
					return
				}
			}
		}
	}()

	// Check for initial WebSocket errors
	if err := ls.wsClient.GetError(); err != nil {
		// Check if this is a normal closure
		if ls.wsClient.IsNormalClosure(err) {
			// Normal closure - don't report as error
			ls.cancel()
			return nil
		}
		// This is an actual error
		ls.cancel()
		return err
	}

	// Wait for context cancellation (stream finished or error)
	<-ls.ctx.Done()
	return nil
}

// formatWebSocketError formats WebSocket errors for better user experience
func formatWebSocketError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Handle specific WebSocket error patterns
	if strings.Contains(errStr, "close 1008") {
		// Extract the reason from the error message
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
		// Default fallback for unknown 1008 errors
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

	// Default error message
	return fmt.Sprintf("Log stream error: %v", err)
}
