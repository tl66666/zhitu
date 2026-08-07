package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
	"strings"
)

// AttachResume 在面试中发送简历快照
// 将指定简历版本的 content 写入 interview.ResumeSnapshot，后续 AI 提问会结合简历内容
func (s *InterviewService) AttachResume(userID, interviewID uint, in *AttachResumeInput) (*models.Interview, error) {
	interview, err := s.Get(userID, interviewID)
	if err != nil {
		return nil, err
	}
	if interview.Status != StatusOngoing {
		return nil, ErrInterviewEnded
	}
	if strings.TrimSpace(interview.ResumeSnapshot) != "" {
		return nil, ErrResumeLocked
	}

	// 1. 校验简历归属
	var resume models.Resume
	if err := s.db.Where("id = ? AND user_id = ?", in.ResumeID, userID).First(&resume).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrResumeNotFound
	} else if err != nil {
		return nil, err
	}

	// 2. 确定版本 ID：优先使用入参，否则用简历当前版本
	versionID := in.VersionID
	if versionID == 0 {
		versionID = resume.CurrentVersionID
	}
	if versionID == 0 {
		return nil, ErrVersionNotFound
	}

	// 3. 拉取版本内容
	var version models.ResumeVersion
	if err := s.db.Where("id = ? AND resume_id = ?", versionID, in.ResumeID).First(&version).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVersionNotFound
	} else if err != nil {
		return nil, err
	}

	// 4. 把简历快照写入面试
	updates := map[string]interface{}{
		"resume_id":         resume.ID,
		"resume_version_id": version.ID,
		"resume_snapshot":   version.Content,
		"resume_name":       resume.Name,
	}
	result := s.db.Model(&models.Interview{}).
		Where("id = ? AND user_id = ? AND status = ?", interviewID, userID, StatusOngoing).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrInterviewEnded
	}
	interview.ResumeSnapshot = version.Content
	interview.ResumeName = resume.Name
	interview.ResumeID = resume.ID
	interview.ResumeVersionID = version.ID
	return interview, nil
}

// summarizeResume 把简历 content JSON 转成可读文本摘要，供 AI prompt 使用
func summarizeResume(contentJSON string) string {
	if strings.TrimSpace(contentJSON) == "" {
		return ""
	}
	var rc ResumeContent
	if err := json.Unmarshal([]byte(contentJSON), &rc); err != nil {
		return ""
	}
	var b strings.Builder

	// 个人信息
	if rc.Personal.Name != "" {
		fmt.Fprintf(&b, "姓名：%s", rc.Personal.Name)
		if rc.Personal.Gender != "" {
			fmt.Fprintf(&b, "｜%s", rc.Personal.Gender)
		}
		if rc.Personal.Age != "" {
			fmt.Fprintf(&b, "｜%s岁", rc.Personal.Age)
		}
		if rc.Personal.City != "" {
			fmt.Fprintf(&b, "｜现居 %s", rc.Personal.City)
		}
		if rc.Personal.Email != "" {
			fmt.Fprintf(&b, "｜%s", rc.Personal.Email)
		}
		if rc.Personal.GitHub != "" {
			fmt.Fprintf(&b, "｜GitHub：%s", rc.Personal.GitHub)
		}
		b.WriteString("\n")
	}

	// 求职意向
	if rc.Intention.Position != "" || rc.Intention.Salary != "" {
		b.WriteString("求职意向：")
		parts := []string{}
		if rc.Intention.Position != "" {
			parts = append(parts, rc.Intention.Position)
		}
		if rc.Intention.City != "" {
			parts = append(parts, "城市："+rc.Intention.City)
		}
		if rc.Intention.Salary != "" {
			parts = append(parts, "期望薪资："+rc.Intention.Salary)
		}
		if rc.Intention.Arrival != "" {
			parts = append(parts, "到岗时间："+rc.Intention.Arrival)
		}
		b.WriteString(strings.Join(parts, "｜"))
		b.WriteString("\n")
	}

	// 教育背景
	if len(rc.Education) > 0 {
		b.WriteString("教育背景：\n")
		for _, e := range rc.Education {
			fmt.Fprintf(&b, "- %s · %s · %s（%s ~ %s）", e.School, e.Major, e.Degree, e.Start, e.End)
			if e.GPA != "" {
				fmt.Fprintf(&b, "｜GPA：%s", e.GPA)
			}
			if e.Courses != "" {
				fmt.Fprintf(&b, "｜主修：%s", e.Courses)
			}
			b.WriteString("\n")
		}
	}

	// 工作经历
	if len(rc.Work) > 0 {
		b.WriteString("工作经历：\n")
		for _, w := range rc.Work {
			fmt.Fprintf(&b, "- %s · %s（%s ~ %s）\n", w.Company, w.Position, w.Start, w.End)
			if w.Description != "" {
				fmt.Fprintf(&b, "  职责：%s\n", w.Description)
			}
			if w.LeaveReason != "" {
				fmt.Fprintf(&b, "  离职原因：%s\n", w.LeaveReason)
			}
		}
	}

	// 项目经历
	if len(rc.Project) > 0 {
		b.WriteString("项目经历：\n")
		for _, p := range rc.Project {
			fmt.Fprintf(&b, "- %s · %s（%s ~ %s）\n", p.Name, p.Role, p.Start, p.End)
			if p.Description != "" {
				fmt.Fprintf(&b, "  描述：%s\n", p.Description)
			}
			if len(p.TechStack) > 0 {
				fmt.Fprintf(&b, "  技术栈：%s\n", strings.Join(p.TechStack, "、"))
			}
			if p.URL != "" {
				fmt.Fprintf(&b, "  链接：%s\n", p.URL)
			}
		}
	}

	// 技能
	if len(rc.Skills) > 0 {
		b.WriteString("技能：\n")
		for _, sk := range rc.Skills {
			fmt.Fprintf(&b, "- %s｜%s", sk.Category, sk.Name)
			if sk.Proficiency != "" {
				fmt.Fprintf(&b, "｜%s", sk.Proficiency)
			}
			b.WriteString("\n")
		}
	}

	// 荣誉
	if len(rc.Honor) > 0 {
		b.WriteString("荣誉奖项：\n")
		for _, h := range rc.Honor {
			fmt.Fprintf(&b, "- %s（%s · %s）\n", h.Name, h.Issuer, h.Date)
			if h.Level != "" {
				fmt.Fprintf(&b, "  级别：%s\n", h.Level)
			}
		}
	}

	// 自定义模块
	for _, c := range rc.Custom {
		if c.Title == "" || c.Content == "" {
			continue
		}
		fmt.Fprintf(&b, "%s：\n%s\n", c.Title, c.Content)
	}

	summary := strings.TrimSpace(b.String())
	runes := []rune(summary)
	if len(runes) > maxResumePromptRunes {
		return string(runes[:maxResumePromptRunes]) + "\n（简历内容已截断）"
	}
	return summary
}
