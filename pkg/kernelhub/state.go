package kernelhub

import (
	"context"
	"errors"
	"sync"

	"k8s.io/klog/v2"
)

var (
	ErrStateNotFound = errors.New("kernel state not found")
)

type MemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]*KernelStateData
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[string]*KernelStateData),
	}
}

func (s *MemoryStateStore) Save(ctx context.Context, sessionID string, data *KernelStateData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[sessionID] = data
	klog.V(4).Infof("Saved state for session %s", sessionID)
	return nil
}

func (s *MemoryStateStore) Load(ctx context.Context, sessionID string) (*KernelStateData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.states[sessionID]
	if !exists {
		return nil, ErrStateNotFound
	}
	return state, nil
}

func (s *MemoryStateStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, sessionID)
	return nil
}
