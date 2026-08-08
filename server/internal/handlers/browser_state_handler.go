package handlers

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zhitu/server/internal/middleware"
	"github.com/zhitu/server/internal/services"
	"github.com/zhitu/server/internal/utils"
)

type BrowserStateHandler struct{ svc *services.BrowserStateService }

func NewBrowserStateHandler(svc *services.BrowserStateService) *BrowserStateHandler {
	return &BrowserStateHandler{svc: svc}
}

func (h *BrowserStateHandler) Get(c *gin.Context) {
	key := c.Param("key")
	value, err := h.svc.Get(c.GetUint(middleware.ContextWorkspaceID), key)
	if err != nil {
		respondBrowserStateErr(c, err)
		return
	}
	utils.OK(c, gin.H{"key": key, "value": json.RawMessage(value)})
}

func (h *BrowserStateHandler) Put(c *gin.Context) {
	var input struct {
		Value json.RawMessage `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	key := c.Param("key")
	if err := h.svc.Put(c.GetUint(middleware.ContextWorkspaceID), key, input.Value); err != nil {
		respondBrowserStateErr(c, err)
		return
	}
	utils.OK(c, gin.H{"key": key})
}

func (h *BrowserStateHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	if err := h.svc.Delete(c.GetUint(middleware.ContextWorkspaceID), key); err != nil {
		respondBrowserStateErr(c, err)
		return
	}
	utils.OK(c, gin.H{"key": key})
}

func respondBrowserStateErr(c *gin.Context, err error) {
	if errors.Is(err, services.ErrBrowserStateKey) || errors.Is(err, services.ErrBrowserStateTooLarge) || errors.Is(err, services.ErrBrowserStateInvalid) {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.InternalError(c, err.Error())
}
