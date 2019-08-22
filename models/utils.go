package models

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"net/url"
	"os"
	"time"
)

//const connStr = "user=vrailean password=5b34b4ccc dbname=prezentr_admin_development host=localhost sslmode=disable"

//os.Getenv("DATABASE_URL")
var (
	DB          *gorm.DB
	CurrentUser User
)

func SetCurrentUser(u *User) {
	CurrentUser = *u
}

func IsCurrentUserPresent() (b bool) {
	b = CurrentUser == User{}
	return
}

func InitDB() *gorm.DB {
	fmt.Println(DB)
	DB, _ = gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if DB == nil {
		panic("db nil")
	}

	DB.DropTableIfExists(&User{}, &Conference{}, &Presentation{}, &Session{}, &Quiz{}, &Answer{}, &QuizAnswer{})
	DB.AutoMigrate(&User{}, &Conference{}, &Presentation{}, &Session{}, &Quiz{}, &Answer{}, &QuizAnswer{})
	DB.Model(&Session{}).AddForeignKey("presentation_id", "presentations(id)", "CASCADE", "CASCADE")
	DB.Model(&Quiz{}).AddForeignKey("presentation_id", "presentations(id)", "CASCADE", "CASCADE")
	DB.Model(&Answer{}).AddForeignKey("quiz_id", "quizzes(id)", "CASCADE", "CASCADE")
	DB.Model(&QuizAnswer{}).AddForeignKey("quiz_id", "quizzes(id)", "CASCADE", "CASCADE")
	//db.Model(&QuizAnswer{}).AddForeignKey("session_id", "sessions(id)", "CASCADE", "CASCADE")
	return DB
}

type Model struct {
	ID        int `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FindUserByTempToken(token string) (user User) {
	user = User{}
	DB.Find(&user, "temporary_token = ?", token)
	SetCurrentUser(&user)
	return
}

func FindUserByPubToken(token string) (user User) {
	user = User{}
	DB.Find(&user, "public_token = ?", token)
	SetCurrentUser(&user)
	return
}

func GenerateTempTokenUrl(tempToken string, baseUrl string) string {
	var buf bytes.Buffer
	buf.WriteString(baseUrl)
	v := url.Values{
		"temporary_token": {tempToken},
	}
	buf.WriteString("?")
	buf.WriteString(v.Encode())
	return buf.String()
}

func GenerateGetMeUrl(token string) string {
	var buf bytes.Buffer
	buf.WriteString("https://api.github.com/user")
	v := url.Values{
		"access_token": {token},
	}
	buf.WriteString("?")
	buf.WriteString(v.Encode())
	return buf.String()
}

func generateAWSLink(fileName string) string {
	var link bytes.Buffer
	link.WriteString("https://presentr-bucket.s3.eu-west-2.amazonaws.com/")
	link.WriteString(fileName)
	return link.String()
}
