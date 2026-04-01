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

package websocket

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	conn             *websocket.Conn
	microserviceUUID string
	execID           string
	done             chan struct{}
	lastMessageType  int // Track the last message type
	err              error
	errMutex         sync.Mutex
	closeOnce        sync.Once
	// Keep-alive fields
	pingTicker   *time.Ticker
	lastPongTime time.Time
	pongMutex    sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(microserviceUUID string) *Client {
	return &Client{
		microserviceUUID: microserviceUUID,
		done:             make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection to the server
func (c *Client) Connect(url string, headers http.Header) error {
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	conn, resp, err := dialer.Dial(url, headers)
	if err != nil {
		// Check if this is a close frame error
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				c.errMutex.Lock()
				c.err = util.NewError(fmt.Sprintf("connection closed by server (code: %d, reason: %s)", closeErr.Code, closeErr.Text))
				c.errMutex.Unlock()
				c.Close()
				return c.err
			}
		}
		// Check if we got a non-200 response
		if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
			c.errMutex.Lock()
			c.err = util.NewError(fmt.Sprintf("failed to establish WebSocket connection: server returned status %d", resp.StatusCode))
			c.errMutex.Unlock()
			c.Close()
			return c.err
		}
		c.errMutex.Lock()
		c.err = util.NewError(fmt.Sprintf("failed to establish WebSocket connection: %v", err))
		c.errMutex.Unlock()
		c.Close()
		return c.err
	}

	c.conn = conn

	// Set up ping/pong handlers
	c.setupPingPong()

	// Start ping loop
	c.startPingLoop()

	return nil
}

// setupPingPong sets up ping/pong handlers for keep-alive
func (c *Client) setupPingPong() {
	// Handle incoming pings from server
	c.conn.SetPingHandler(func(appData string) error {
		// Respond with pong immediately
		return c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(DefaultPongTimeout*time.Millisecond))
	})

	// Handle incoming pongs from server
	c.conn.SetPongHandler(func(appData string) error {
		c.pongMutex.Lock()
		c.lastPongTime = time.Now()
		c.pongMutex.Unlock()

		// Extend read deadline
		return c.conn.SetReadDeadline(time.Now().Add(DefaultPingInterval * 2 * time.Millisecond))
	})

	// Set initial read deadline
	c.conn.SetReadDeadline(time.Now().Add(DefaultPingInterval * 2 * time.Millisecond))
}

// startPingLoop starts the periodic ping sending loop
func (c *Client) startPingLoop() {
	c.pingTicker = time.NewTicker(DefaultPingInterval * time.Millisecond)

	go func() {
		defer c.pingTicker.Stop()

		for {
			select {
			case <-c.pingTicker.C:
				if !c.IsConnected() {
					return
				}

				// Send ping to server
				if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(DefaultPongTimeout*time.Millisecond)); err != nil {
					c.handlePingError(err)
					return
				}

				// Check pong timeout in background
				go c.checkPongTimeout()

			case <-c.done:
				return
			}
		}
	}()
}

// checkPongTimeout checks if we received a pong response within timeout
func (c *Client) checkPongTimeout() {
	time.Sleep(DefaultPongTimeout * time.Millisecond)

	// Check if connection is still active
	if !c.IsConnected() {
		return // Connection already closed, no need to check timeout
	}

	c.pongMutex.RLock()
	lastPong := c.lastPongTime
	c.pongMutex.RUnlock()

	if time.Since(lastPong) > DefaultPongTimeout*time.Millisecond {
		c.errMutex.Lock()
		c.err = util.NewError("pong timeout - server not responding")
		c.errMutex.Unlock()
		c.Close()
	}
}

// handlePingError handles ping sending errors
func (c *Client) handlePingError(err error) {
	// Check if this is a normal closure error
	if c.IsNormalClosure(err) {
		// Normal closure - don't treat as error
		c.Close()
		return
	}

	// This is an actual ping error
	c.errMutex.Lock()
	c.err = util.NewError(fmt.Sprintf("ping failed: %v", err))
	c.errMutex.Unlock()
	c.Close()
}

// SendMessage sends a message to the server
func (c *Client) SendMessage(msg *Message) error {
	if c.conn == nil {
		return util.NewError("not connected")
	}

	// Set session-specific fields
	msg.MicroserviceUUID = c.microserviceUUID
	msg.ExecID = c.execID

	data, err := msg.Encode()
	if err != nil {
		return util.NewError(fmt.Sprintf("failed to encode message: %v", err))
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// ReadMessage reads a message from the server
func (c *Client) ReadMessage() (*Message, error) {
	if c.conn == nil {
		return nil, util.NewError("not connected")
	}

	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		// Check if this is a normal closure
		if c.IsNormalClosure(err) {
			// Normal closure - don't treat as error, just close gracefully
			c.Close()
			return nil, nil // Return nil to indicate normal termination
		}

		// This is an actual error - format and store it
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
			// Extract close code and reason if available
			if closeErr, ok := err.(*websocket.CloseError); ok {
				err = util.NewError(fmt.Sprintf("connection closed by server (code: %d, reason: %s)", closeErr.Code, closeErr.Text))
			}
		} else {
			err = util.NewError(fmt.Sprintf("failed to read message: %v", err))
		}
		// Store the error and close connection
		c.errMutex.Lock()
		c.err = err
		c.errMutex.Unlock()
		c.Close()
		return nil, err
	}

	// Store the message type
	c.lastMessageType = messageType

	// All messages are now MessagePack encoded
	msg, err := Decode(data)
	if err != nil {
		// Store the error and close connection
		c.errMutex.Lock()
		c.err = err
		c.errMutex.Unlock()
		c.Close()
		return nil, err
	}

	// Update session-specific fields
	c.execID = msg.ExecID

	// Update last pong time (activity detected)
	c.pongMutex.Lock()
	c.lastPongTime = time.Now()
	c.pongMutex.Unlock()

	// Extend read deadline
	c.conn.SetReadDeadline(time.Now().Add(DefaultPingInterval * 2 * time.Millisecond))

	return msg, nil
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		// Stop ping ticker
		if c.pingTicker != nil {
			c.pingTicker.Stop()
		}

		if c.conn != nil {
			// Send close message before closing
			closeMsg := NewMessage(MessageTypeClose, nil, c.microserviceUUID, c.execID)
			_ = c.SendMessage(closeMsg)
			closeErr = c.conn.Close()
			c.conn = nil
		}
		close(c.done)
	})
	return closeErr
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	return c.conn != nil
}

// GetDone returns the done channel
func (c *Client) GetDone() <-chan struct{} {
	return c.done
}

// SetExecID sets the execution ID for the client
func (c *Client) SetExecID(execID string) {
	c.execID = execID
}

// GetExecID returns the current execution ID
func (c *Client) GetExecID() string {
	return c.execID
}

// GetMicroserviceUUID returns the microservice UUID
func (c *Client) GetMicroserviceUUID() string {
	return c.microserviceUUID
}

// GetError returns the last error that occurred
func (c *Client) GetError() error {
	c.errMutex.Lock()
	defer c.errMutex.Unlock()
	return c.err
}

// IsNormalClosure checks if the error represents a normal session closure
func (c *Client) IsNormalClosure(err error) bool {
	if err == nil {
		return false
	}

	// Check for WebSocket close errors with normal codes
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return true
	}

	// Check for "use of closed network connection" during intentional close
	if closeErr, ok := err.(*websocket.CloseError); ok {
		return closeErr.Code == websocket.CloseNormalClosure ||
			closeErr.Code == websocket.CloseGoingAway ||
			closeErr.Code == websocket.CloseNoStatusReceived
	}

	// Check for network errors that occur during intentional close
	errStr := err.Error()
	if strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "broken pipe") {
		// These are expected when we intentionally close the connection
		return true
	}

	return false
}
