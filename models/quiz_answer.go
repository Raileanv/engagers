package models

import "github.com/jinzhu/gorm"

type QuizAnswer struct {
	gorm.Model
	SessionID uint
	QuizID    uint
	AnswerID  uint
	UserID    uint
}
