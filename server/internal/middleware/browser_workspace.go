package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhitu/server/internal/models"
	"github.com/zhitu/server/internal/utils"
	"gorm.io/gorm"
)

const (
	HeaderBrowserToken   = "X-Browser-Token"
	ContextAccountUserID = "account_user_id"
	ContextWorkspaceID   = "workspace_id"
)

// BrowserWorkspaceScope 将业务请求的 user_id 映射为当前浏览器的工作区 ID。
// 该中间件必须放在 JWTAuth 之后。
func BrowserWorkspaceScope(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountUserID := c.GetUint(ContextUserID)
		rawToken := strings.ToLower(strings.TrimSpace(c.GetHeader(HeaderBrowserToken)))
		decoded, err := hex.DecodeString(rawToken)
		if err != nil || len(decoded) != 32 || len(rawToken) != 64 {
			utils.BadRequest(c, "missing or invalid X-Browser-Token header")
			c.Abort()
			return
		}

		digest := sha256.Sum256(decoded)
		tokenHash := hex.EncodeToString(digest[:])
		now := time.Now()
		workspace := models.BrowserWorkspace{
			AccountUserID: accountUserID,
			TokenHash:     tokenHash,
			LastSeenAt:    now,
		}
		if err := db.Where("account_user_id = ? AND token_hash = ?", accountUserID, tokenHash).
			FirstOrCreate(&workspace).Error; err != nil {
			utils.InternalError(c, "resolve browser workspace failed")
			c.Abort()
			return
		}
		_ = db.Model(&workspace).Update("last_seen_at", now).Error

		c.Set(ContextAccountUserID, accountUserID)
		c.Set(ContextWorkspaceID, workspace.ID)
		c.Set(ContextUserID, workspace.ID)
		c.Next()
	}
}
