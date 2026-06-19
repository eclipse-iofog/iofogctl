package websocket

// Message types for WebSocket communication
const (
	MessageTypeStdin      uint8 = 0
	MessageTypeStdout     uint8 = 1
	MessageTypeStderr     uint8 = 2
	MessageTypeControl    uint8 = 3
	MessageTypeClose      uint8 = 4
	MessageTypeActivation uint8 = 5
	MessageTypeLogLine    uint8 = 6
	MessageTypeLogStart   uint8 = 7
	MessageTypeLogStop    uint8 = 8
	MessageTypeLogError   uint8 = 9
)

// WebSocket configuration constants
const (
	DefaultPingInterval    = 30000       // 30 seconds
	DefaultPongTimeout     = 10000       // 10 seconds
	DefaultMaxPayload      = 1024 * 1024 // 1MB
	DefaultSessionTimeout  = 300000      // 5 minutes
	DefaultCleanupInterval = 60000       // 1 minute
	DefaultMaxConnections  = 10
)
