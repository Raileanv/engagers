package models

import "github.com/jinzhu/gorm"

type QuizAnswer struct {
	gorm.Model
	SessionID uint `json:"session_id"`
	QuizID    uint `json:"quiz_id"`
	AnswerID  uint `json:"answer_id"`
	UserID    uint `json:"user_id"`
}
