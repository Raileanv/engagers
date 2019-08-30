package models

import (
	"github.com/jinzhu/gorm"
	_"gopkg.in/go-playground/validator.v9"
)


type Quiz struct {
	gorm.Model
	Question       string       `json:"question"`
	PresentationID uint         `json:"presentation_id" validate:"required"`
	Type           string       `json:"type" validate:"required,oneof=input select"`
	Answers        []Answer     `json:"answers" gorm:"ForeignKey:QuizID"`
	QuizAnswers    []QuizAnswer `gorm:"ForeignKey:QuizID"`
}

type Answer struct {
	gorm.Model
	Answer  string `json:"answer"`
	Correct bool   `json:"correct"`
	QuizID  uint
}
