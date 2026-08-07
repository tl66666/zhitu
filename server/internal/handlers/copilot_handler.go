package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/zhitu/server/internal/middleware"
	"github.com/zhitu/server/internal/services"
	"github.com/zhitu/server/internal/utils"
)

// CopilotHandler exposes the browser-backed, task-oriented resume assistant.
type CopilotHandler struct {
	svc *services.ResumeCopilotService
}

func NewCopilotHandler(svc *services.ResumeCopilotService) *CopilotHandler {
	return &CopilotHandler{svc: svc}
}

func (h *CopilotHandler) Chat(c *gin.Context) {
	userID := c.GetUint(middleware.ContextUserID)
	var in services.CopilotInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	h.setSSEHeaders(c)
	flusher, ok := c.Writer.(interface{ Flush() })
	if !ok {
		utils.InternalError(c, "streaming not supported")
		return
	}
	status, _ := json.Marshal(map[string]string{"type": "status", "message": "正在读取简历上下文并规划任务"})
	fmt.Fprintf(c.Writer, "data: %s\n\n", status)
	flusher.Flush()

	result, err := h.svc.Chat(c.Request.Context(), userID, &in)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"type": "error", "message": err.Error()})
		fmt.Fprintf(c.Writer, "data: %s\n\n", errData)
		flusher.Flush()
		return
	}
	done, _ := json.Marshal(map[string]interface{}{
		"type":           "done",
		"message":        map[string]string{"role": "assistant", "content": result.Reply},
		"result":         result,
		"proposals":      result.Proposals,
		"memory_summary": result.MemorySummary,
	})
	fmt.Fprintf(c.Writer, "data: %s\n\n", done)
	flusher.Flush()
}

func (h *CopilotHandler) Apply(c *gin.Context) {
	userID := c.GetUint(middleware.ContextUserID)
	var in services.CopilotApplyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	version, err := h.svc.Apply(c.Request.Context(), userID, &in)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrResumeNotFound), errors.Is(err, services.ErrVersionNotFound):
			utils.NotFound(c, err.Error())
		case errors.Is(err, services.ErrCopilotContentTooLong), errors.Is(err, services.ErrCopilotInvalidContent), errors.Is(err, services.ErrCopilotProjectRange):
			utils.BadRequest(c, err.Error())
		case errors.Is(err, services.ErrCopilotResumeChanged):
			utils.Conflict(c, err.Error())
		case errors.Is(err, services.ErrLLMNotConfigured):
			utils.InternalError(c, err.Error())
		default:
			utils.Conflict(c, err.Error())
		}
		return
	}
	utils.OKWithMsg(c, "copilot proposal applied as a new version", version)
}

func (h *CopilotHandler) setSSEHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
}
