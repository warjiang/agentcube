package kernelhub

import (
	"context"
	"time"
)

type SessionState string

const (
	SessionStatePending  SessionState = "pending"
	SessionStateRunning SessionState = "running"
	SessionStateIdle    SessionState = "idle"
	SessionStateBusy    SessionState = "busy"
	SessionStateStopping SessionState = "stopping"
)

type KernelState string

const (
	KernelStateStarting   KernelState = "starting"
	KernelStateIdle       KernelState = "idle"
	KernelStateBusy       KernelState = "busy"
	KernelStateRestarting KernelState = "restarting"
	KernelStateDead       KernelState = "dead"
)

type ExecuteResult struct {
	Status    string `json:"status"`
	Count     int    `json:"count"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ErrorType string `json:"error_type,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

type KernelStateData struct {
	SessionID      string                 `json:"session_id"`
	Namespace      map[string]interface{} `json:"namespace"`
	ExecutionCount int                   `json:"execution_count"`
	CreatedAt      time.Time             `json:"created_at"`
	SavedAt        time.Time             `json:"saved_at"`
}

type Session struct {
	ID             string
	Kernel        *KernelProcess
	State         SessionState
	CreatedAt     time.Time
	LastActiveAt  time.Time
	ExecutionCount int
}

type SessionInfo struct {
	SessionID      string        `json:"session_id"`
	State         SessionState  `json:"state"`
	CreatedAt     time.Time     `json:"created_at"`
	LastActiveAt  time.Time     `json:"last_active_at"`
	ExecutionCount int           `json:"execution_count"`
	HasKernel     bool          `json:"has_kernel"`
	KernelInfo    *KernelInfo   `json:"kernel_info,omitempty"`
}

type KernelInfo struct {
	ID             string      `json:"id"`
	SessionID      string      `json:"session_id"`
	State          KernelState `json:"state"`
	MemoryUsage    int64       `json:"memory_usage"`
	MemoryLimit    int64       `json:"memory_limit"`
	StartedAt      time.Time   `json:"started_at"`
	LastActiveAt   time.Time   `json:"last_active_at"`
	ExecutionCount int         `json:"execution_count"`
}

type StateStore interface {
	Save(ctx context.Context, sessionID string, data *KernelStateData) error
	Load(ctx context.Context, sessionID string) (*KernelStateData, error)
	Delete(ctx context.Context, sessionID string) error
}

type Config struct {
	Port          int
	Workspace     string
	MemoryLimit   int64
	MaxSessions   int
	SessionTTL    time.Duration
	StateStore    string
	RedisAddr     string
	RedisPassword string
}
