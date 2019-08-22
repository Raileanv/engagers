package models

import (
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

type Conference struct {
	gorm.Model
	UserId      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
}

func mapRequestToConference(request *http.Request, conference *Conference) {
	layoutISO := "2006-01-02T15:04:05"
	startAt, _ := time.Parse(layoutISO, request.FormValue("start_at"))
	endAt, _ := time.Parse(layoutISO, request.FormValue("end_at"))

	conference.StartAt = startAt
	conference.EndAt = endAt
	conference.UserId = CurrentUser.ID
	conference.Title = request.FormValue("title")
	conference.Description = request.FormValue("description")
	conference.Thumbnail = request.FormValue("thumbnail")
}

func CreateConferenceHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	var conference Conference
	mapRequestToConference(r, &conference)

	DB.Create(&conference)
}