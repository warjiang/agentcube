package kernelhub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrKernelNotFound  = errors.New("kernel not found")
	ErrMaxSessions     = errors.New("max sessions reached")
)

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session

	config   ManagerConfig
	executor *Executor
}

type ManagerConfig struct {
	DefaultMemoryLimit int64
	DefaultTimeout     time.Duration
	MaxSessions        int
	SessionTTL         time.Duration
	stateStore         StateStore
}

func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		config: ManagerConfig{
			DefaultMemoryLimit: 512 * 1024 * 1024,
			DefaultTimeout:     60 * time.Second,
			MaxSessions:        100,
			SessionTTL:         24 * time.Hour,
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	m.executor = NewExecutor(m)

	return m
}

type ManagerOption func(*Manager)

func WithMemoryLimit(limit int64) ManagerOption {
	return func(m *Manager) {
		m.config.DefaultMemoryLimit = limit
	}
}

func WithMaxSessions(max int) ManagerOption {
	return func(m *Manager) {
		m.config.MaxSessions = max
	}
}

func WithSessionTTL(ttl time.Duration) ManagerOption {
	return func(m *Manager) {
		m.config.SessionTTL = ttl
	}
}

func WithStateStore(store StateStore) ManagerOption {
	return func(m *Manager) {
		m.config.stateStore = store
	}
}

func (m *Manager) CreateSession(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		return session, nil
	}

	if len(m.sessions) >= m.config.MaxSessions {
		return nil, ErrMaxSessions
	}

	session := &Session{
		ID:        sessionID,
		State:     SessionStatePending,
		CreatedAt: time.Now(),
	}

	m.sessions[sessionID] = session
	klog.Infof("Session created: %s", sessionID)

	return session, nil
}

func (m *Manager) StartKernel(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.Lock()
	session, exists := m.sessions[sessionID]
	if !exists {
		m.mu.Unlock()
		return nil, ErrSessionNotFound
	}

	if session.Kernel != nil {
		m.mu.Unlock()
		return session, nil
	}

	kernel := &KernelProcess{
		ID:          sessionID,
		MemoryLimit: m.config.DefaultMemoryLimit,
		Timeout:     m.config.DefaultTimeout,
	}

	if err := kernel.Start(); err != nil {
		m.mu.Unlock()
		return nil, fmt.Errorf("start kernel: %w", err)
	}

	session.Kernel = kernel
	session.State = SessionStateIdle

	if m.config.stateStore != nil {
		state, err := m.config.stateStore.Load(ctx, sessionID)
		if err == nil && state != nil {
			kernel.RestoreNamespace(state.Namespace)
			kernel.ExecutionCount = state.ExecutionCount
			klog.Infof("Restored state for session %s", sessionID)
		}
	}

	m.mu.Unlock()

	klog.Infof("Kernel started for session %s", sessionID)
	return session, nil
}

func (m *Manager) ExecuteCode(ctx context.Context, sessionID string, code string, timeout time.Duration) (*ExecuteResult, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	if session.Kernel == nil {
		if _, err := m.StartKernel(ctx, sessionID); err != nil {
			return nil, err
		}
	}

	session.LastActiveAt = time.Now()
	session.ExecutionCount++

	result := session.Kernel.Execute(code, timeout)

	return result, nil
}

func (m *Manager) StopKernel(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.Kernel == nil {
		return nil
	}

	if m.config.stateStore != nil {
		state := &KernelStateData{
			SessionID:      sessionID,
			Namespace:     session.Kernel.GetNamespace(),
			ExecutionCount: session.Kernel.ExecutionCount,
			SavedAt:       time.Now(),
		}
		if err := m.config.stateStore.Save(ctx, sessionID, state); err != nil {
			klog.Warningf("Failed to save state: %v", err)
		}
	}

	session.Kernel.Stop()
	session.Kernel = nil
	session.State = SessionStatePending

	klog.Infof("Kernel stopped for session %s", sessionID)
	return nil
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.Kernel != nil {
		session.Kernel.Stop()
	}

	if m.config.stateStore != nil {
		m.config.stateStore.Delete(ctx, sessionID)
	}

	delete(m.sessions, sessionID)
	klog.Infof("Session deleted: %s", sessionID)

	return nil
}

func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}

func (m *Manager) CleanupIdle(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if session.State == SessionStateIdle &&
			now.Sub(session.LastActiveAt) > m.config.SessionTTL {

			if session.Kernel != nil {
				session.Kernel.Stop()
			}
			delete(m.sessions, id)
			klog.Infof("Cleaned up idle session: %s", id)
		}
	}
}

func (m *Manager) GetKernel(sessionID string) *KernelProcess {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil
	}

	return session.Kernel
}
