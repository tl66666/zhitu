package models

import "time"

// BrowserWorkspace 将同一登录账号下的数据按浏览器隔离。
// 浏览器原始令牌不落库，仅保存其 SHA-256 摘要。
type BrowserWorkspace struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AccountUserID uint      `gorm:"not null;uniqueIndex:idx_browser_workspace_account_token" json:"account_user_id"`
	TokenHash     string    `gorm:"size:64;not null;uniqueIndex:idx_browser_workspace_account_token" json:"-"`
	LastSeenAt    time.Time `gorm:"not null" json:"last_seen_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (BrowserWorkspace) TableName() string { return "browser_workspaces" }

// BrowserState 保存工作区级的轻量 JSON 状态，例如 Copilot 会话历史。
type BrowserState struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WorkspaceID uint      `gorm:"not null;uniqueIndex:idx_browser_state_workspace_key" json:"workspace_id"`
	StateKey    string    `gorm:"size:80;not null;uniqueIndex:idx_browser_state_workspace_key" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (BrowserState) TableName() string { return "browser_states" }
