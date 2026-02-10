package executor

import (
	"context"
	"fmt"
	"time"
)

const sessionKeyPrefix = "exec_session:"
const sessionTTL = 1 * time.Hour

// SessionManager manages execution session state in Redis (or similar store).
type SessionManager struct {
	get  func(ctx context.Context, key string) (string, error)
	set  func(ctx context.Context, key string, value string, ttl time.Duration) error
	del  func(ctx context.Context, keys ...string) error
}

// NewSessionManager creates a session manager using the provided get/set/del functions (e.g. Redis).
func NewSessionManager(get func(context.Context, string) (string, error), set func(context.Context, string, string, time.Duration) error, del func(context.Context, ...string) error) *SessionManager {
	return &SessionManager{get: get, set: set, del: del}
}

func sessionKey(notebookID string) string {
	return sessionKeyPrefix + notebookID
}

// GetOrCreateSession returns the current session code (accumulated so far). If none, returns empty string.
func (m *SessionManager) GetOrCreateSession(ctx context.Context, notebookID string) (string, error) {
	key := sessionKey(notebookID)
	s, err := m.get(ctx, key)
	if err != nil || s == "" {
		return "", nil
	}
	return s, nil
}

// SetSession saves the full session code (call after successful execution).
func (m *SessionManager) SetSession(ctx context.Context, notebookID string, fullCode string) error {
	return m.set(ctx, sessionKey(notebookID), fullCode, sessionTTL)
}

// ClearSession removes the session for the notebook.
func (m *SessionManager) ClearSession(ctx context.Context, notebookID string) error {
	return m.del(ctx, sessionKey(notebookID))
}

// SessionInfo returns a simple info string for GET /session (e.g. "has_session:true").
func (m *SessionManager) SessionInfo(ctx context.Context, notebookID string) (string, error) {
	s, err := m.get(ctx, sessionKey(notebookID))
	if err != nil {
		return "", err
	}
	if s == "" {
		return "has_session:false", nil
	}
	return fmt.Sprintf("has_session:true,length:%d", len(s)), nil
}
