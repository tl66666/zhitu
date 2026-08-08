package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
)

// 投递相关错误
var (
	ErrDeliveryNotFound      = errors.New("delivery not found")
	ErrRoundNotFound         = errors.New("delivery round not found")
	ErrFeedbackNotFound      = errors.New("delivery feedback not found")
	ErrResumeVersionNotFound = errors.New("resume version not found")
	ErrInvalidStatus         = errors.New("invalid delivery status")
	ErrInvalidTransition     = errors.New("invalid status transition")
)

// 合法的状态集合
var validStatuses = map[string]bool{
	models.DeliveryStatusPending:      true,
	models.DeliveryStatusWrittenTest:  true,
	models.DeliveryStatusInterview:    true,
	models.DeliveryStatusWaitingOffer: true,
	models.DeliveryStatusOffer:        true,
	models.DeliveryStatusRejected:     true,
}

// 合法的状态流转：from -> [允许到达的 to]
var validTransitions = map[string]map[string]bool{
	models.DeliveryStatusPending: {
		models.DeliveryStatusWrittenTest: true,
		models.DeliveryStatusInterview:   true,
		models.DeliveryStatusRejected:    true,
	},
	models.DeliveryStatusWrittenTest: {
		models.DeliveryStatusInterview:    true,
		models.DeliveryStatusWaitingOffer: true,
		models.DeliveryStatusRejected:     true,
	},
	models.DeliveryStatusInterview: {
		models.DeliveryStatusWaitingOffer: true,
		models.DeliveryStatusRejected:     true,
	},
	models.DeliveryStatusWaitingOffer: {
		models.DeliveryStatusOffer:    true,
		models.DeliveryStatusRejected: true,
	},
	models.DeliveryStatusOffer: {
		models.DeliveryStatusRejected: true, // 放弃 Offer
	},
	models.DeliveryStatusRejected: {
		models.DeliveryStatusInterview: true, // 被拒后复活
	},
}

// DeliveryService 投递看板业务逻辑
type DeliveryService struct {
	db *gorm.DB
}

// NewDeliveryService 构造 DeliveryService
func NewDeliveryService(db *gorm.DB) *DeliveryService {
	return &DeliveryService{db: db}
}

// ---------- 投递主表 CRUD ----------

// CreateDeliveryInput 创建投递入参
type CreateDeliveryInput struct {
	Company     string `json:"company" binding:"required"`
	Position    string `json:"position" binding:"required"`
	Channel     string `json:"channel" binding:"required"`
	ApplyDate   string `json:"apply_date" binding:"required"`
	Priority    string `json:"priority"`
	JDText      string `json:"jd_text"`
	ResumeVerID uint   `json:"resume_version_id"`
	HRContact   string `json:"hr_contact"`
	NextStep    string `json:"next_step"`
	Remark      string `json:"remark"`
}

// List 列出用户投递记录，支持按 status / channel 筛选
func (s *DeliveryService) List(userID uint, status, channel string) ([]models.Delivery, error) {
	q := s.db.Where("user_id = ?", userID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if channel != "" {
		q = q.Where("channel = ?", channel)
	}
	var list []models.Delivery
	err := q.Order("apply_date DESC, id DESC").Find(&list).Error
	return list, err
}

// Get 获取单条投递
func (s *DeliveryService) Get(userID, id uint) (*models.Delivery, error) {
	var d models.Delivery
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDeliveryNotFound
	}
	return &d, err
}

// GetDetail 获取投递详情（含轮次与反馈）
func (s *DeliveryService) GetDetail(userID, id uint) (*models.Delivery, []models.DeliveryRound, []models.DeliveryFeedback, error) {
	d, err := s.Get(userID, id)
	if err != nil {
		return nil, nil, nil, err
	}
	var rounds []models.DeliveryRound
	s.db.Where("delivery_id = ?", id).Order("interview_time ASC, id ASC").Find(&rounds)
	var feedbacks []models.DeliveryFeedback
	s.db.Where("delivery_id = ?", id).Order("contact_time DESC, id DESC").Find(&feedbacks)
	return d, rounds, feedbacks, nil
}

// Create 创建投递记录
func (s *DeliveryService) Create(userID uint, in *CreateDeliveryInput) (*models.Delivery, error) {
	if !validChannel(in.Channel) {
		return nil, fmt.Errorf("invalid channel: %s", in.Channel)
	}
	if in.Priority == "" {
		in.Priority = models.PriorityMedium
	}
	if !validPriority(in.Priority) {
		return nil, fmt.Errorf("invalid priority: %s", in.Priority)
	}
	if in.ResumeVerID != 0 {
		belongs, err := s.resumeVersionBelongsToUser(userID, in.ResumeVerID)
		if err != nil {
			return nil, err
		}
		if !belongs {
			return nil, ErrResumeVersionNotFound
		}
	}
	d := &models.Delivery{
		UserID:      userID,
		Company:     in.Company,
		Position:    in.Position,
		Channel:     in.Channel,
		ApplyDate:   in.ApplyDate,
		Status:      models.DeliveryStatusPending,
		Priority:    in.Priority,
		JDText:      in.JDText,
		ResumeVerID: in.ResumeVerID,
		HRContact:   in.HRContact,
		NextStep:    in.NextStep,
		Remark:      in.Remark,
	}
	if err := s.db.Create(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// Update 更新投递记录的可变字段
func (s *DeliveryService) Update(userID, id uint, updates map[string]interface{}) error {
	allowed := map[string]bool{
		"company": true, "position": true, "channel": true, "apply_date": true,
		"priority": true, "jd_text": true, "resume_version_id": true,
		"hr_contact": true, "next_step": true, "offer_detail": true, "remark": true,
	}
	filtered := map[string]interface{}{}
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}
	if v, ok := filtered["channel"]; ok {
		if !validChannel(fmt.Sprintf("%v", v)) {
			return fmt.Errorf("invalid channel: %v", v)
		}
	}
	if v, ok := filtered["priority"]; ok {
		if !validPriority(fmt.Sprintf("%v", v)) {
			return fmt.Errorf("invalid priority: %v", v)
		}
	}
	if v, ok := filtered["resume_version_id"]; ok {
		versionID, ok := deliveryUpdateUint(v)
		if !ok || versionID == 0 {
			return ErrResumeVersionNotFound
		}
		belongs, err := s.resumeVersionBelongsToUser(userID, versionID)
		if err != nil {
			return err
		}
		if !belongs {
			return ErrResumeVersionNotFound
		}
		filtered["resume_version_id"] = versionID
	}
	if len(filtered) == 0 {
		return nil
	}
	result := s.db.Model(&models.Delivery{}).Where("id = ? AND user_id = ?", id, userID).Updates(filtered)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeliveryNotFound
	}
	return nil
}

func (s *DeliveryService) resumeVersionBelongsToUser(userID, versionID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.ResumeVersion{}).
		Joins("JOIN resumes ON resumes.id = resume_versions.resume_id").
		Where("resume_versions.id = ? AND resumes.user_id = ?", versionID, userID).
		Count(&count).Error
	return count > 0, err
}

func deliveryUpdateUint(value interface{}) (uint, bool) {
	switch v := value.(type) {
	case uint:
		return v, true
	case uint8:
		return uint(v), true
	case uint16:
		return uint(v), true
	case uint32:
		return uint(v), true
	case uint64:
		return uint(v), uint64(uint(v)) == v
	case int:
		return uint(v), v >= 0
	case int8:
		return uint(v), v >= 0
	case int16:
		return uint(v), v >= 0
	case int32:
		return uint(v), v >= 0
	case int64:
		return uint(v), v >= 0
	case float64:
		return uint(v), v >= 0 && v == float64(uint(v))
	case float32:
		return uint(v), v >= 0 && v == float32(uint(v))
	case json.Number:
		n, err := v.Int64()
		return uint(n), err == nil && n >= 0
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		return uint(n), err == nil && uint64(uint(n)) == n
	default:
		return 0, false
	}
}

// Delete 删除投递记录（级联删除轮次与反馈）
func (s *DeliveryService) Delete(userID, id uint) error {
	if _, err := s.Get(userID, id); err != nil {
		return err
	}
	if err := s.db.Where("delivery_id = ?", id).Delete(&models.DeliveryRound{}).Error; err != nil {
		return err
	}
	if err := s.db.Where("delivery_id = ?", id).Delete(&models.DeliveryFeedback{}).Error; err != nil {
		return err
	}
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Delivery{}).Error
}

// ChangeStatus 变更投递状态，校验流转合法性
func (s *DeliveryService) ChangeStatus(userID, id uint, to string) (*models.Delivery, error) {
	if !validStatuses[to] {
		return nil, ErrInvalidStatus
	}
	d, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if d.Status == to {
		return d, nil
	}
	allowed, ok := validTransitions[d.Status]
	if !ok || !allowed[to] {
		return nil, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, d.Status, to)
	}
	if err := s.db.Model(d).Update("status", to).Error; err != nil {
		return nil, err
	}
	d.Status = to
	return d, nil
}

// ---------- 面试轮次 ----------

// ListRounds 列出投递的所有面试轮次
func (s *DeliveryService) ListRounds(userID, deliveryID uint) ([]models.DeliveryRound, error) {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return nil, err
	}
	var list []models.DeliveryRound
	err := s.db.Where("delivery_id = ?", deliveryID).Order("interview_time ASC, id ASC").Find(&list).Error
	return list, err
}

// CreateRound 新增面试轮次
func (s *DeliveryService) CreateRound(userID, deliveryID uint, r *models.DeliveryRound) (*models.DeliveryRound, error) {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return nil, err
	}
	if !validRoundType(r.RoundType) {
		return nil, fmt.Errorf("invalid round_type: %s", r.RoundType)
	}
	if r.Format != "" && !validFormat(r.Format) {
		return nil, fmt.Errorf("invalid format: %s", r.Format)
	}
	if r.Result == "" {
		r.Result = models.RoundResultPending
	}
	if !validRoundResult(r.Result) {
		return nil, fmt.Errorf("invalid result: %s", r.Result)
	}
	r.ID = 0
	r.DeliveryID = deliveryID
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// UpdateRound 更新面试轮次
func (s *DeliveryService) UpdateRound(userID, deliveryID, roundID uint, updates map[string]interface{}) error {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return err
	}
	if v, ok := updates["round_type"]; ok {
		if !validRoundType(fmt.Sprintf("%v", v)) {
			return fmt.Errorf("invalid round_type: %v", v)
		}
	}
	if v, ok := updates["format"]; ok {
		s := fmt.Sprintf("%v", v)
		if s != "" && !validFormat(s) {
			return fmt.Errorf("invalid format: %v", v)
		}
	}
	if v, ok := updates["result"]; ok {
		if !validRoundResult(fmt.Sprintf("%v", v)) {
			return fmt.Errorf("invalid result: %v", v)
		}
	}
	result := s.db.Model(&models.DeliveryRound{}).
		Where("id = ? AND delivery_id = ?", roundID, deliveryID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRoundNotFound
	}
	return nil
}

// DeleteRound 删除面试轮次
func (s *DeliveryService) DeleteRound(userID, deliveryID, roundID uint) error {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return err
	}
	result := s.db.Where("id = ? AND delivery_id = ?", roundID, deliveryID).Delete(&models.DeliveryRound{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRoundNotFound
	}
	return nil
}

// ---------- HR 反馈 ----------

// ListFeedbacks 列出投递的 HR 反馈记录
func (s *DeliveryService) ListFeedbacks(userID, deliveryID uint) ([]models.DeliveryFeedback, error) {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return nil, err
	}
	var list []models.DeliveryFeedback
	err := s.db.Where("delivery_id = ?", deliveryID).Order("contact_time DESC, id DESC").Find(&list).Error
	return list, err
}

// CreateFeedback 新增 HR 反馈记录
func (s *DeliveryService) CreateFeedback(userID, deliveryID uint, f *models.DeliveryFeedback) (*models.DeliveryFeedback, error) {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return nil, err
	}
	f.ID = 0
	f.DeliveryID = deliveryID
	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	// 反馈中的 next_action 若非空，同步写回投递记录的 next_step 字段（便于看板展示最新动态）
	if f.NextAction != "" {
		payload := map[string]interface{}{"todo": f.NextAction}
		b, _ := json.Marshal(payload)
		s.db.Model(&models.Delivery{}).Where("id = ?", deliveryID).Update("next_step", string(b))
	}
	return f, nil
}

// DeleteFeedback 删除 HR 反馈记录
func (s *DeliveryService) DeleteFeedback(userID, deliveryID, feedbackID uint) error {
	if _, err := s.Get(userID, deliveryID); err != nil {
		return err
	}
	result := s.db.Where("id = ? AND delivery_id = ?", feedbackID, deliveryID).Delete(&models.DeliveryFeedback{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFeedbackNotFound
	}
	return nil
}

// ---------- 统计与漏斗 ----------

// DeliveryStats 投递统计
type DeliveryStats struct {
	Total      int64 `json:"total"`       // 总投递数
	InProgress int64 `json:"in_progress"` // 进行中（pending + written_test + interview + waiting_offer）
	OfferCount int64 `json:"offer_count"` // 已获 Offer
	Rejected   int64 `json:"rejected"`    // 已拒绝
}

// GetStats 统计用户投递概览
func (s *DeliveryService) GetStats(userID uint) (*DeliveryStats, error) {
	st := &DeliveryStats{}
	s.db.Model(&models.Delivery{}).Where("user_id = ?", userID).Count(&st.Total)
	s.db.Model(&models.Delivery{}).Where("user_id = ? AND status IN ?", userID,
		[]string{models.DeliveryStatusPending, models.DeliveryStatusWrittenTest,
			models.DeliveryStatusInterview, models.DeliveryStatusWaitingOffer}).Count(&st.InProgress)
	s.db.Model(&models.Delivery{}).Where("user_id = ? AND status = ?", userID, models.DeliveryStatusOffer).Count(&st.OfferCount)
	s.db.Model(&models.Delivery{}).Where("user_id = ? AND status = ?", userID, models.DeliveryStatusRejected).Count(&st.Rejected)
	return st, nil
}

// DeliveryFunnel 投递漏斗
type DeliveryFunnel struct {
	Applied         int64   `json:"applied"`           // 投递数
	WrittenTestPass int64   `json:"written_test_pass"` // 笔试通过数
	FirstPass       int64   `json:"first_pass"`        // 一面通过数
	SecondPass      int64   `json:"second_pass"`       // 二面通过数
	OfferCount      int64   `json:"offer_count"`       // Offer 数
	WrittenTestRate float64 `json:"written_test_rate"` // 笔试通过率
	FirstRate       float64 `json:"first_rate"`        // 一面通过率
	SecondRate      float64 `json:"second_rate"`       // 二面通过率
	OfferRate       float64 `json:"offer_rate"`        // 总 Offer 获得率
}

// GetFunnel 计算投递漏斗
// 通过统计到达各阶段的投递数（status 已到达该阶段或更后阶段）
func (s *DeliveryService) GetFunnel(userID uint) (*DeliveryFunnel, error) {
	f := &DeliveryFunnel{}

	// 投递数
	s.db.Model(&models.Delivery{}).Where("user_id = ?", userID).Count(&f.Applied)

	// 笔试通过数：有 pass 的笔试轮次（用轮次口径更精确，rejected 可能发生在笔试前）
	s.db.Model(&models.DeliveryRound{}).
		Joins("JOIN deliveries ON deliveries.id = delivery_rounds.delivery_id").
		Where("deliveries.user_id = ? AND delivery_rounds.round_type = ? AND delivery_rounds.result = ?",
			userID, models.RoundWrittenTest, models.RoundResultPass).
		Distinct("delivery_rounds.delivery_id").
		Count(&f.WrittenTestPass)

	// 一面通过数：有一面 pass 的轮次
	s.db.Model(&models.DeliveryRound{}).
		Joins("JOIN deliveries ON deliveries.id = delivery_rounds.delivery_id").
		Where("deliveries.user_id = ? AND delivery_rounds.round_type = ? AND delivery_rounds.result = ?",
			userID, models.RoundFirstTech, models.RoundResultPass).
		Distinct("delivery_rounds.delivery_id").
		Count(&f.FirstPass)

	// 二面通过数
	s.db.Model(&models.DeliveryRound{}).
		Joins("JOIN deliveries ON deliveries.id = delivery_rounds.delivery_id").
		Where("deliveries.user_id = ? AND delivery_rounds.round_type = ? AND delivery_rounds.result = ?",
			userID, models.RoundSecondTech, models.RoundResultPass).
		Distinct("delivery_rounds.delivery_id").
		Count(&f.SecondPass)

	// Offer 数
	s.db.Model(&models.Delivery{}).Where("user_id = ? AND status = ?", userID, models.DeliveryStatusOffer).Count(&f.OfferCount)

	// 转化率
	if f.Applied > 0 {
		f.WrittenTestRate = pct(f.WrittenTestPass, f.Applied)
		f.OfferRate = pct(f.OfferCount, f.Applied)
	}
	if f.WrittenTestPass > 0 {
		f.FirstRate = pct(f.FirstPass, f.WrittenTestPass)
	}
	if f.FirstPass > 0 {
		f.SecondRate = pct(f.SecondPass, f.FirstPass)
	}
	return f, nil
}

// pct 计算百分比，保留 2 位小数
func pct(num, denom int64) float64 {
	if denom == 0 {
		return 0
	}
	return float64(num) / float64(denom) * 100
}

// ---------- 校验工具 ----------

func validChannel(c string) bool {
	switch c {
	case models.ChannelBoss, models.ChannelOfficial, models.ChannelReferral,
		models.ChannelCampus, models.ChannelHeadhunt, models.ChannelOther:
		return true
	}
	return false
}

func validPriority(p string) bool {
	switch p {
	case models.PriorityHigh, models.PriorityMedium, models.PriorityLow:
		return true
	}
	return false
}

func validRoundType(t string) bool {
	switch t {
	case models.RoundWrittenTest, models.RoundFirstTech, models.RoundSecondTech,
		models.RoundThirdTech, models.RoundCross, models.RoundHR,
		models.RoundAdditional, models.RoundFinal:
		return true
	}
	return false
}

func validFormat(f string) bool {
	switch f {
	case models.FormatOnsite, models.FormatVideo, models.FormatPhone, "":
		return true
	}
	return false
}

func validRoundResult(r string) bool {
	switch r {
	case models.RoundResultPass, models.RoundResultPending, models.RoundResultReject:
		return true
	}
	return false
}
