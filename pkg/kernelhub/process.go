package kernelhub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type KernelProcess struct {
	ID             string
	MemoryLimit    int64
	Timeout        time.Duration
	State          KernelState
	StartedAt      time.Time
	LastActiveAt   time.Time
	ExecutionCount int

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	mu       sync.RWMutex
	requests map[string]chan *ExecuteResult
}

type KernelMessage struct {
	Type    string          `json:"type"`
	MsgID   string          `json:"msg_id"`
	Content json.RawMessage `json:"content"`
}

type ExecuteContent struct {
	Code    string `json:"code"`
	Silent  bool   `json:"silent,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type ExecuteReplyContent struct {
	Status string `json:"status"`
	Count  int    `json:"count,omitempty"`
	Error  string `json:"error,omitempty"`
}

type StreamContent struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func NewKernelProcess(id string, memoryLimit int64, timeout time.Duration) *KernelProcess {
	return &KernelProcess{
		ID:          id,
		MemoryLimit: memoryLimit,
		Timeout:     timeout,
		State:       KernelStateStarting,
		StartedAt:   time.Now(),
	}
}

func (kp *KernelProcess) Start() error {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	kp.requests = make(map[string]chan *ExecuteResult)

	cmd := exec.Command("python3", "-m", "agentcube.kernelhub",
		"--session-id", kp.ID,
		"--memory-limit", strconv.FormatInt(kp.MemoryLimit, 10))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	kp.stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	kp.stdout = stdout
	kp.scanner = bufio.NewScanner(stdout)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	kp.cmd = cmd
	kp.State = KernelStateIdle

	go kp.readResponses()

	return kp.waitForReady(5 * time.Second)
}

func (kp *KernelProcess) waitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("kernel ready timeout")
		case <-ticker.C:
			kp.mu.RLock()
			state := kp.State
			kp.mu.RUnlock()
			if state == KernelStateIdle {
				return nil
			}
		}
	}
}

func (kp *KernelProcess) readResponses() {
	for kp.scanner.Scan() {
		line := kp.scanner.Bytes()

		var msg KernelMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		kp.mu.Lock()
		ch, exists := kp.requests[msg.MsgID]
		delete(kp.requests, msg.MsgID)
		kp.mu.Unlock()

		if !exists {
			continue
		}

		switch msg.Type {
		case "execute_reply":
			var content ExecuteReplyContent
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				ch <- &ExecuteResult{
					Status:  "error",
					ErrorMsg: fmt.Sprintf("parse error: %v", err),
				}
			} else {
				kp.mu.Lock()
				if content.Status == "ok" {
					kp.ExecutionCount = content.Count
				}
				kp.mu.Unlock()

				ch <- &ExecuteResult{
					Status: content.Status,
					Count:  content.Count,
					Stderr: content.Error,
				}
			}

		case "stream":
			var content StreamContent
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				continue
			}

			if content.Name == "stdout" {
				ch <- &ExecuteResult{
					Status: "ok",
					Count:  kp.ExecutionCount,
					Stdout: content.Text,
				}
			} else {
				ch <- &ExecuteResult{
					Status: "ok",
					Count:  kp.ExecutionCount,
					Stderr: content.Text,
				}
			}
		}
	}
}

func (kp *KernelProcess) Execute(code string, timeout time.Duration) *ExecuteResult {
	kp.mu.Lock()

	if kp.State == KernelStateDead {
		kp.mu.Unlock()
		return &ExecuteResult{
			Status:  "error",
			ErrorMsg: "kernel is dead",
		}
	}

	kp.State = KernelStateBusy
	kp.LastActiveAt = time.Now()

	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	resultCh := make(chan *ExecuteResult, 10)
	kp.requests[msgID] = resultCh

	content := ExecuteContent{
		Code:    code,
		Timeout: int(timeout.Seconds()),
	}
	contentBytes, _ := json.Marshal(content)

	msg := KernelMessage{
		Type:    "execute",
		MsgID:   msgID,
		Content: contentBytes,
	}

	msgBytes, _ := json.Marshal(msg)
	kp.stdin.Write(append(msgBytes, '\n'))

	kp.mu.Unlock()

	var response *ExecuteResult
	for {
		select {
		case resp := <-resultCh:
			if response == nil {
				response = resp
			} else {
				response.Stdout += resp.Stdout
				response.Stderr += resp.Stderr
			}
		case <-time.After(timeout + 5*time.Second):
			kp.mu.Lock()
			delete(kp.requests, msgID)
			kp.mu.Unlock()

			kp.Interrupt()
			return &ExecuteResult{
				Status:  "timeout",
				ErrorMsg: fmt.Sprintf("execution timeout after %v", timeout),
			}
		}

		kp.mu.RLock()
		_, exists := kp.requests[msgID]
		kp.mu.RUnlock()

		if !exists && len(resultCh) == 0 {
			break
		}
	}

	kp.mu.Lock()
	kp.State = KernelStateIdle
	kp.mu.Unlock()

	return response
}

func (kp *KernelProcess) Interrupt() error {
	kp.mu.RLock()
	cmd := kp.cmd
	kp.mu.RUnlock()

	if cmd != nil && cmd.Process != nil {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}
	return nil
}

func (kp *KernelProcess) Stop() error {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	kp.State = KernelStateDead

	if kp.cmd != nil && kp.cmd.Process != nil {
		syscall.Kill(-kp.cmd.Process.Pid, syscall.SIGKILL)
		kp.cmd.Wait()
	}

	if kp.stdin != nil {
		kp.stdin.Close()
	}

	return nil
}

func (kp *KernelProcess) Info() *KernelInfo {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	return &KernelInfo{
		ID:             kp.ID,
		SessionID:      kp.ID,
		State:          kp.State,
		MemoryUsage:    kp.getMemoryUsage(),
		MemoryLimit:    kp.MemoryLimit,
		StartedAt:      kp.StartedAt,
		LastActiveAt:   kp.LastActiveAt,
		ExecutionCount: kp.ExecutionCount,
	}
}

func (kp *KernelProcess) getMemoryUsage() int64 {
	if kp.cmd == nil || kp.cmd.Process == nil {
		return 0
	}

	statusFile := fmt.Sprintf("/proc/%d/status", kp.cmd.Process.Pid)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, _ := strconv.ParseInt(parts[1], 10, 64)
				return kb * 1024
			}
		}
	}

	return 0
}

func (kp *KernelProcess) GetNamespace() map[string]interface{} {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	msg := KernelMessage{
		Type:  "get_namespace",
		MsgID: msgID,
	}

	msgBytes, _ := json.Marshal(msg)
	kp.stdin.Write(append(msgBytes, '\n'))

	for kp.scanner.Scan() {
		line := kp.scanner.Bytes()

		var resp KernelMessage
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		if resp.MsgID == msgID && resp.Type == "namespace" {
			var ns map[string]interface{}
			if err := json.Unmarshal(resp.Content, &ns); err != nil {
				return nil
			}
			return ns
		}
	}

	return nil
}

func (kp *KernelProcess) RestoreNamespace(ns map[string]interface{}) error {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	content, _ := json.Marshal(ns)
	msg := KernelMessage{
		Type:    "restore_namespace",
		MsgID:   msgID,
		Content: content,
	}

	msgBytes, _ := json.Marshal(msg)
	kp.stdin.Write(append(msgBytes, '\n'))

	return nil
}

func (kp *KernelProcess) StopStdin() error {
	return nil
}
