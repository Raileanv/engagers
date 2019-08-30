package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// belongs to presentation
// has info about start / end date
// has info about conference
type Session struct {
	gorm.Model
	StartAt        time.Time `json:"start_at" gorm:"not null"`
	EndAt          time.Time `json:"end_at" gorm:"not null"`
	PresentationID uint `json:"presentation_id" validate:"required"`
	ConferenceID   uint `json:"conference_id"`
	QuizAnswers    []QuizAnswer `json:"quiz_answers" gorm:"ForeignKey:SessionID"`
	TvToken        string `json:"tv_token"`
}
