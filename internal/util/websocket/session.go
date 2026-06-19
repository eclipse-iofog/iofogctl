package websocket

import (
	"sync"
	"time"
)

// Session represents a WebSocket session between a user and an agent
type Session struct {
	ExecID           string
	MicroserviceUUID string
	LastActivity     time.Time
	mu               sync.RWMutex
}

// NewSession creates a new Session with the given execID and microserviceUUID
func NewSession(execID, microserviceUUID string) *Session {
	return &Session{
		ExecID:           execID,
		MicroserviceUUID: microserviceUUID,
		LastActivity:     time.Now(),
	}
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActivity = time.Now()
}

// IsExpired checks if the session has expired based on the timeout
func (s *Session) IsExpired(timeout time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.LastActivity) > timeout
}

// SessionManager manages WebSocket sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager creates a new SessionManager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// AddSession adds a new session to the manager
func (sm *SessionManager) AddSession(session *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ExecID] = session
}

// GetSession retrieves a session by its execID
func (sm *SessionManager) GetSession(execID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[execID]
}

// RemoveSession removes a session by its execID
func (sm *SessionManager) RemoveSession(execID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, execID)
}

// CleanupExpiredSessions removes all expired sessions
func (sm *SessionManager) CleanupExpiredSessions(timeout time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for execID, session := range sm.sessions {
		if session.IsExpired(timeout) {
			delete(sm.sessions, execID)
		}
	}
}

// GetActiveSessions returns the number of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
