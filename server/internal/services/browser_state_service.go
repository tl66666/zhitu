package services

import (
	"encoding/json"
	"errors"

	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MaxBrowserStateBytes = 1024 * 1024

var (
	ErrBrowserStateKey      = errors.New("unsupported browser state key")
	ErrBrowserStateTooLarge = errors.New("browser state is too large")
	ErrBrowserStateInvalid  = errors.New("browser state must be valid JSON")
)

type BrowserStateService struct{ db *gorm.DB }

func NewBrowserStateService(db *gorm.DB) *BrowserStateService {
	return &BrowserStateService{db: db}
}

func validateBrowserState(key string, value []byte) error {
	if key != "copilot_sessions" {
		return ErrBrowserStateKey
	}
	if len(value) > MaxBrowserStateBytes {
		return ErrBrowserStateTooLarge
	}
	if !json.Valid(value) {
		return ErrBrowserStateInvalid
	}
	return nil
}

func (s *BrowserStateService) Get(workspaceID uint, key string) ([]byte, error) {
	if key != "copilot_sessions" {
		return nil, ErrBrowserStateKey
	}
	var state models.BrowserState
	err := s.db.Where("workspace_id = ? AND state_key = ?", workspaceID, key).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []byte("[]"), nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(state.Value), nil
}

func (s *BrowserStateService) Put(workspaceID uint, key string, value []byte) error {
	if err := validateBrowserState(key, value); err != nil {
		return err
	}
	state := models.BrowserState{WorkspaceID: workspaceID, StateKey: key, Value: string(value)}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "workspace_id"}, {Name: "state_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&state).Error
}

func (s *BrowserStateService) Delete(workspaceID uint, key string) error {
	if key != "copilot_sessions" {
		return ErrBrowserStateKey
	}
	return s.db.Where("workspace_id = ? AND state_key = ?", workspaceID, key).
		Delete(&models.BrowserState{}).Error
}
