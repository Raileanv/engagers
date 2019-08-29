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
	StartAt        time.Time `gorm:"not null"`
	EndAt          time.Time `gorm:"not null"`
	PresentationID uint
	ConferenceID   uint
	QuizAnswers    []QuizAnswer `gorm:"ForeignKey:SessionID"`
	TvToken        string
}
