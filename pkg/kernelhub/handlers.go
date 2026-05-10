package kernelhub

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	manager *Manager
}

func NewHandlers(m *Manager) *Handlers {
	return &Handlers{manager: m}
}

func (h *Handlers) CreateSession(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required"})
		return
	}

	session, err := h.manager.CreateSession(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"state":      session.State,
		"created_at": session.CreatedAt,
	})
}

func (h *Handlers) GetSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.manager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":       session.ID,
		"state":           session.State,
		"created_at":      session.CreatedAt,
		"last_active_at":  session.LastActiveAt,
		"execution_count": session.ExecutionCount,
		"has_kernel":      session.Kernel != nil,
	})
}

func (h *Handlers) DeleteSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	if err := h.manager.DeleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *Handlers) ListSessions(c *gin.Context) {
	sessions := h.manager.ListSessions()

	result := make([]gin.H, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, gin.H{
			"session_id":       s.ID,
			"state":           s.State,
			"created_at":      s.CreatedAt,
			"last_active_at":  s.LastActiveAt,
			"execution_count": s.ExecutionCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": result})
}

func (h *Handlers) StartKernel(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.manager.StartKernel(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.ID,
		"state":      session.State,
	})
}

func (h *Handlers) ExecuteCode(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		Code    string `json:"code" binding:"required"`
		Timeout string `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeout := 60 * time.Second
	if req.Timeout != "" {
		var err error
		timeout, err = time.ParseDuration(req.Timeout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timeout"})
			return
		}
	}

	result, err := h.manager.ExecuteCode(c.Request.Context(), sessionID, req.Code, timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handlers) StopKernel(c *gin.Context) {
	sessionID := c.Param("session_id")

	if err := h.manager.StopKernel(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

func (h *Handlers) InterruptKernel(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.manager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if session.Kernel != nil {
		session.Kernel.Interrupt()
	}

	c.JSON(http.StatusOK, gin.H{"status": "interrupted"})
}

func (h *Handlers) RestartKernel(c *gin.Context) {
	sessionID := c.Param("session_id")

	h.manager.StopKernel(c.Request.Context(), sessionID)

	session, err := h.manager.StartKernel(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": session.ID,
		"state":      session.State,
	})
}

func (h *Handlers) GetState(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.manager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if session.Kernel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "kernel not started"})
		return
	}

	ns := session.Kernel.GetNamespace()

	c.JSON(http.StatusOK, gin.H{
		"session_id":       sessionID,
		"namespace":       ns,
		"execution_count": session.Kernel.ExecutionCount,
	})
}

func (h *Handlers) RestoreState(c *gin.Context) {
	sessionID := c.Param("session_id")

	var req struct {
		Namespace       map[string]interface{} `json:"namespace"`
		ExecutionCount int                   `json:"execution_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.manager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if session.Kernel == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kernel not started"})
		return
	}

	session.Kernel.RestoreNamespace(req.Namespace)
	session.Kernel.ExecutionCount = req.ExecutionCount

	c.JSON(http.StatusOK, gin.H{"status": "restored"})
}

func (h *Handlers) GetKernelInfo(c *gin.Context) {
	sessionID := c.Param("session_id")

	session, err := h.manager.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if session.Kernel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "kernel not started"})
		return
	}

	c.JSON(http.StatusOK, session.Kernel.Info())
}
