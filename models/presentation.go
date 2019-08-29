package models

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSessionPackage "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/globalsign/mgo/bson"
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Presentation struct {
	gorm.Model
	ConferenceId   uint   `json:"conference_id"`
	UserId         uint   `json:"user_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Thumbnail      string `json:"thumbnail"`
	AttachmentLink string
	Session        []Session `gorm:"ForeignKey:PresentationID"`
	Quiz           []Quiz    `gorm:"ForeignKey:PresentationID"`
}

type Presentations []Presentation

var (
	awsSession, _ = awsSessionPackage.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ID"),
			os.Getenv("AWS_SECRET"),
			""), // token can be left blank for now
	})
)

func mapRequestToSession(request *http.Request, session *Session) {
	layoutISO := "2006-01-02T15:04:05"
	startAt, _ := time.Parse(layoutISO, request.FormValue("start_at"))
	endAt, _ := time.Parse(layoutISO, request.FormValue("end_at"))
	conferenceId, _ := strconv.ParseInt(request.FormValue("conference_id"), 10, 32)

	session.ConferenceID = uint(conferenceId)
	session.StartAt = startAt
	session.EndAt = endAt
}

func GetPresentationsHandler(w http.ResponseWriter, r *http.Request) {
	var presentations Presentations
	DB.Table("presentations").Scan(&presentations)
	jsonPresentations, _ := json.Marshal(presentations)
	fmt.Fprint(w, string(jsonPresentations))
}

func GetPresentationHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, _ := strconv.ParseInt(params["presentation_id"], 10, 32)

	presentation := Presentation{}
	DB.First(&presentation, id)
	if presentation.ID != 0 {
		pr, _ := json.Marshal(presentation)

		w.Write(pr)
		return
	}
	fmt.Fprintf(w, "No presentation with id: %d ", id)
}

func CreatePresentationHandler(w http.ResponseWriter, r *http.Request) {

	var presentation Presentation
	var session Session
	var quizes []Quiz

	presentation.UserId = CurrentUser.ID

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if part.FormName() == "title" {
			data, _ := ioutil.ReadAll(part)
			presentation.Title = string(data)
		}
		if part.FormName() == "description" {
			data, _ := ioutil.ReadAll(part)
			presentation.Description = string(data)
		}
		if part.FormName() == "thumbnail" {
			data, _ := ioutil.ReadAll(part)
			presentation.Thumbnail = string(data)
		}
		if part.FormName() == "conference_id" {
			data, _ := ioutil.ReadAll(part)
			conference_id, _ := strconv.ParseInt(string(data), 10, 32)
			session.ConferenceID = uint(conference_id)
		}
		if part.FormName() == "start_at" {
			layoutISO := "2006-01-02T15:04:05"
			data, _ := ioutil.ReadAll(part)
			startAt, _ := time.Parse(layoutISO, string(data))
			session.StartAt = startAt
		}
		if part.FormName() == "end_at" {
			layoutISO := "2006-01-02T15:04:05"
			data, _ := ioutil.ReadAll(part)
			endAt, _ := time.Parse(layoutISO, string(data))
			session.StartAt = endAt
		}

		if part.FormName() == "presentation_attachment" {
			file, _ := ioutil.ReadAll(part)

			if file != nil {

				fileName, err := uploadFileToS3(awsSession, file, part.FileName(), binary.Size(file))

				if err != nil {
					_, _ = fmt.Fprintf(w, "Could not upload file \n", err)
					http.Error(w, "Could not upload file", http.StatusNotFound)
				}
				presentation.AttachmentLink = generateAWSLink(fileName)
			}
		}

		if part.FormName() == "quizes" {
			jsonDecoder := json.NewDecoder(part)

			err = jsonDecoder.Decode(&quizes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	presentation.Quiz = quizes
	presentation.Session = append(presentation.Session, session)

	DB.Create(&presentation)

	jsonPresentation, err := json.Marshal(presentation)

	w.Write(jsonPresentation)
}

func GetPresentationSessionsHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, _ := strconv.ParseInt(params["presentation_id"], 10, 32)
	var presentation Presentation

	DB.First(&presentation, id)

	var sessions []Session
	DB.Model(&presentation).Related(&sessions)

	jsonSessions, _ := json.Marshal(sessions)
	w.Write(jsonSessions)
}

func PostAddSessionToPresentation(w http.ResponseWriter, r *http.Request, params martini.Params) {
	var session Session

	id, _ := strconv.ParseInt(params["presentation_id"], 10, 32)

	mapRequestToSession(r, &session)

	session.PresentationID = uint(id)
	DB.Save(&session)

	w.WriteHeader(http.StatusOK)
}

func PostAddQuizToPresentation(w http.ResponseWriter, r *http.Request, params martini.Params) {
	id, _ := strconv.ParseInt(params["presentation_id"], 10, 32)
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	quizes := []Quiz{}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if part.FormName() == "quizes" {
			jsonDecoder := json.NewDecoder(part)
			fmt.Println("decoder ", jsonDecoder)
			err = jsonDecoder.Decode(&quizes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	for _, n := range quizes {
		n.PresentationID = uint(id)
		DB.Save(n)
	}
	jsn, _ := json.Marshal(quizes)
	w.Write(jsn)
}

func uploadFileToS3(s *awsSessionPackage.Session, file []byte, filename string, size int) (string, error) {
	// create a unique file name for the file
	tempFileName := "presentations/" + bson.NewObjectId().Hex() + filename

	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String("presentr-bucket"),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:                 bytes.NewReader(file),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(file)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}
