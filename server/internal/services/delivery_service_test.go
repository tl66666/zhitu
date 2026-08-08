package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

func TestDeliveryDeleteScopesRecordAndChildren(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&models.Delivery{}, &models.DeliveryRound{}, &models.DeliveryFeedback{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	target := models.Delivery{UserID: 7, Company: "A", Position: "工程师", Channel: models.ChannelOfficial, ApplyDate: "2026-08-08"}
	other := models.Delivery{UserID: 8, Company: "B", Position: "工程师", Channel: models.ChannelOfficial, ApplyDate: "2026-08-08"}
	if err := db.Create(&target).Error; err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other: %v", err)
	}
	if err := db.Create(&models.DeliveryRound{DeliveryID: target.ID, RoundType: models.RoundFirstTech}).Error; err != nil {
		t.Fatalf("create round: %v", err)
	}
	if err := db.Create(&models.DeliveryFeedback{DeliveryID: target.ID, ContactTime: "2026-08-08 10:00"}).Error; err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	service := NewDeliveryService(db)
	if err := service.Delete(7, target.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var targetCount, otherCount, roundCount, feedbackCount int64
	db.Model(&models.Delivery{}).Where("id = ?", target.ID).Count(&targetCount)
	db.Model(&models.Delivery{}).Where("id = ?", other.ID).Count(&otherCount)
	db.Model(&models.DeliveryRound{}).Where("delivery_id = ?", target.ID).Count(&roundCount)
	db.Model(&models.DeliveryFeedback{}).Where("delivery_id = ?", target.ID).Count(&feedbackCount)
	if targetCount != 0 || roundCount != 0 || feedbackCount != 0 {
		t.Fatalf("target remains: delivery=%d rounds=%d feedbacks=%d", targetCount, roundCount, feedbackCount)
	}
	if otherCount != 1 {
		t.Fatalf("other user's delivery count = %d, want 1", otherCount)
	}
}

func TestDeliveryCreateScopesResumeVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&models.Resume{}, &models.ResumeVersion{}, &models.Delivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	resume := models.Resume{UserID: 7, Name: "后端简历"}
	if err := db.Create(&resume).Error; err != nil {
		t.Fatalf("create resume: %v", err)
	}
	version := models.ResumeVersion{ResumeID: resume.ID, VersionLabel: "v1.0", Content: "{}"}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("create resume version: %v", err)
	}

	service := NewDeliveryService(db)
	input := &CreateDeliveryInput{
		Company: "示例公司", Position: "后端工程师", Channel: models.ChannelOfficial,
		ApplyDate: "2026-08-08", ResumeVerID: version.ID,
	}
	if _, err := service.Create(7, input); err != nil {
		t.Fatalf("owner create: %v", err)
	}
	if _, err := service.Create(8, input); !errors.Is(err, ErrResumeVersionNotFound) {
		t.Fatalf("cross-user create error = %v, want ErrResumeVersionNotFound", err)
	}
}
