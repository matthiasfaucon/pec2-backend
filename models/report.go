package models

import "time"

type ReportReason string

const (
	DISLIKE          ReportReason = "DISLIKE"
	HARASSMENT       ReportReason = "HARASSMENT"
	SELF_HARM        ReportReason = "SELF_HARM"
	VIOLENCE         ReportReason = "VIOLENCE"
	RESTRICTED_ITEMS ReportReason = "RESTRICTED_ITEMS"
	NUDITY           ReportReason = "NUDITY"
	SCAM             ReportReason = "SCAM"
	MISINFORMATION   ReportReason = "MISINFORMATION"
	ILLEGAL_CONTENT  ReportReason = "ILLEGAL_CONTENT"
)

type Report struct {
	ID         string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID     string       `json:"postId" gorm:"column:post_id"`
	ReportedBy string       `json:"reportedBy" gorm:"column:reported_by"`
	Reason     ReportReason `json:"reason" gorm:"column:reason"`
	CreatedAt  time.Time    `json:"createdAt"`
}

type ReportCreate struct {
	Reason ReportReason `json:"reason" binding:"required"`
}

func (Report) TableName() string {
	return "reports"
}
