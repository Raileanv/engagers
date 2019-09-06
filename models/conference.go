package models

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	//"github.com/jinzhu/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"gopkg.in/go-playground/validator.v9"
)

type Conference struct {
	Model
	UserId      uint      `json:"user_id"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
}

type Conferences []Conference
var validate *validator.Validate

func CreateConferenceHandler(w http.ResponseWriter, r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var conference Conference
	conference.UserId = CurrentUser.ID
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if part.FormName() == "title" {
			data, err := ioutil.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			conference.Title = string(data)
		}
		if part.FormName() == "thumbnail" {
			file, _ := ioutil.ReadAll(part)

			if file != nil {

				fileName, err := uploadFileToS3(awsSession, file, part.FileName(), "conf_thumbnail", conference.Title, binary.Size(file))

				if err != nil {
					_, _ = fmt.Fprintf(w, "Could not upload file \n", err)
					http.Error(w, "Could not upload file", http.StatusNotFound)
				}
				conference.Thumbnail = generateAWSLink(fileName)
			}
		}
		if part.FormName() == "description" {
			data, err := ioutil.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			conference.Description = string(data)
		}
		if part.FormName() == "start_at" {
			data, err := ioutil.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			startAt, err := time.Parse(time.RFC3339, string(data))
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			conference.StartAt = startAt
		}
		if part.FormName() == "end_at" {
			data, err := ioutil.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			endAt, err := time.Parse(time.RFC3339, string(data))
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			conference.StartAt = endAt
		}
	}

	validate = validator.New()
	err = validate.Struct(conference)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	DB.Create(&conference)
	jsn, _ := json.Marshal(conference)
	_, _ = w.Write(jsn)
}

func GetConferencesHandler(w http.ResponseWriter, r *http.Request) {
	var conferences Conferences
	DB.Table("conferences").Scan(&conferences)
	jsonConferences, _ := json.Marshal(conferences)
	fmt.Fprint(w, string(jsonConferences))
}

func GetConferenceHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, _ := strconv.ParseInt(params["conference_id"], 10, 32)

	conference := Conference{}
	DB.First(&conference, id)
	if conference.ID != 0 {
		cf, _ := json.Marshal(conference)

		w.Write(cf)
		return
	}
	fmt.Fprintf(w, "No conference with id: %d ", id)
}
