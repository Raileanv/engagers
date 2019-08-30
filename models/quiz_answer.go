package models

type QuizAnswer struct {
	Model
	SessionID uint `json:"session_id"`
	QuizID    uint `json:"quiz_id"`
	AnswerID  uint `json:"answer_id"`
	UserID    uint `json:"user_id"`
}
