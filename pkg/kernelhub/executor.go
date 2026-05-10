package kernelhub

import (
	"context"
	"sync"
	"time"
)

type ExecuteRequest struct {
	Code     string
	Timeout  time.Duration
	ResultCh chan *ExecuteResult
}

type Executor struct {
	mu      sync.Mutex
	queues  map[string]*RequestQueue
	manager *Manager
}

type RequestQueue struct {
	mu      sync.Mutex
	cond    *sync.Cond
	pending []*ExecuteRequest
	closed  bool
}

func NewRequestQueue() *RequestQueue {
	q := &RequestQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *RequestQueue) Enqueue(req *ExecuteRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		req.ResultCh <- &ExecuteResult{
			Status:  "error",
			ErrorMsg: "queue closed",
		}
		return
	}

	q.pending = append(q.pending, req)
	q.cond.Signal()
}

func (q *RequestQueue) Dequeue() *ExecuteRequest {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.pending) == 0 && !q.closed {
		q.cond.Wait()
	}

	if q.closed {
		return nil
	}

	req := q.pending[0]
	q.pending = q.pending[1:]
	return req
}

func (q *RequestQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closed = true
	q.cond.Broadcast()

	for _, req := range q.pending {
		req.ResultCh <- &ExecuteResult{
			Status:  "error",
			ErrorMsg: "queue closed",
		}
	}
	q.pending = nil
}

func NewExecutor(manager *Manager) *Executor {
	return &Executor{
		queues:  make(map[string]*RequestQueue),
		manager: manager,
	}
}

func (e *Executor) Execute(ctx context.Context, sessionID string, req *ExecuteRequest) *ExecuteResult {
	queue := e.getOrCreateQueue(sessionID)
	queue.Enqueue(req)

	select {
	case result := <-req.ResultCh:
		return result
	case <-ctx.Done():
		return &ExecuteResult{
			Status:  "error",
			ErrorMsg: "context cancelled",
		}
	}
}

func (e *Executor) getOrCreateQueue(sessionID string) *RequestQueue {
	e.mu.Lock()
	defer e.mu.Unlock()

	if queue, exists := e.queues[sessionID]; exists {
		return queue
	}

	queue := NewRequestQueue()
	e.queues[sessionID] = queue
	go e.worker(sessionID, queue)

	return queue
}

func (e *Executor) worker(sessionID string, queue *RequestQueue) {
	for {
		req := queue.Dequeue()
		if req == nil {
			return
		}

		kernel := e.manager.GetKernel(sessionID)
		if kernel == nil {
			req.ResultCh <- &ExecuteResult{
				Status:  "error",
				ErrorMsg: "kernel not found",
			}
			continue
		}

		result := kernel.Execute(req.Code, req.Timeout)
		req.ResultCh <- result
	}
}

func (e *Executor) CloseQueue(sessionID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if queue, exists := e.queues[sessionID]; exists {
		queue.Close()
		delete(e.queues, sessionID)
	}
}
